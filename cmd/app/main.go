package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/config"
	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/handlers"
	storage "github.com/UserNameShouldBeHere/VK-doodle-jump/internal/repository/tarantool"
	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/services"
)

func main() {
	appConfig, err := config.Parse("./cmd/app/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address:  fmt.Sprintf("%s:%d", appConfig.Server.DB.Host, appConfig.Server.DB.Port),
		User:     appConfig.Server.DB.User,
		Password: appConfig.Server.DB.Password,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		fmt.Println("Connection refused:", err)
		return
	}

	loggerConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	sugarLogger := logger.Sugar()

	storageCtx, storageCancel := context.WithCancel(context.Background())
	usersStorage, err := storage.NewUsersStorage(storageCtx, conn)
	if err != nil {
		log.Fatal(err)
	}
	adminShopStorage, err := storage.NewAdminShopStorage(conn)
	if err != nil {
		log.Fatal(err)
	}
	shopStorage, err := storage.NewShopStorage(conn)
	if err != nil {
		log.Fatal(err)
	}
	authStorage, err := storage.NewAuthStorage(conn)
	if err != nil {
		log.Fatal(err)
	}

	adminShopStorage.FillAdmins(context.Background(), appConfig.Server.DB.Admins)

	usersService, err := services.NewUsersService(usersStorage, sugarLogger)
	if err != nil {
		log.Fatal(err)
	}
	adminShopService, err := services.NewAdminShopService(adminShopStorage, sugarLogger)
	if err != nil {
		log.Fatal(err)
	}
	shopService, err := services.NewShopService(shopStorage, sugarLogger)
	if err != nil {
		log.Fatal(err)
	}
	authService, err := services.NewAuthService(authStorage, sugarLogger)
	if err != nil {
		log.Fatal(err)
	}

	gameHandler, err := handlers.NewGameHandler(usersService, sugarLogger)
	if err != nil {
		log.Fatalf("Failed to init game handler: %v", err)
	}
	profileHandler, err := handlers.NewProfileHandler(usersService, sugarLogger)
	if err != nil {
		log.Fatalf("Failed to init profile handler: %v", err)
	}
	adminShopHandler, err := handlers.NewAdminShopHandler(adminShopService, sugarLogger)
	if err != nil {
		log.Fatalf("Failed to init shop handler: %v", err)
	}
	shopHandler, err := handlers.NewShopHandler(shopService, sugarLogger)
	if err != nil {
		log.Fatalf("Failed to init shop handler: %v", err)
	}
	authHandler, err := handlers.NewAuthHandler(authService, sugarLogger)
	if err != nil {
		log.Fatalf("Failed to init auth handler: %v", err)
	}
	middlewareHandler, err := handlers.NewMiddlewareHandler(
		fmt.Sprintf("%s:%d", appConfig.Client.Host, appConfig.Client.Port),
		authService,
		logger.Sugar())
	if err != nil {
		log.Fatalf("Failed to init middleware handler: %v", err)
	}

	router := initRouter(
		authHandler,
		gameHandler,
		profileHandler,
		adminShopHandler,
		shopHandler,
		middlewareHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", appConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		sigInt := make(chan os.Signal, 1)
		signal.Notify(sigInt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigInt
		storageCancel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Failed to stop server gracefully: %v", err)
		}
	}()

	log.Printf("Starting server at http://%s:%d", appConfig.Server.Host, appConfig.Server.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}

	<-stopped

	log.Println("Server stopped")
}

func initRouter(
	authHandler *handlers.AuthHandler,
	gameHandler *handlers.GameHandler,
	profileHandler *handlers.ProfileHandler,
	adminShopHandler *handlers.AdminShopHandler,
	shopHandler *handlers.ShopHandler,
	middlewareHandler *handlers.MiddlewareHandler) *mux.Router {

	router := mux.NewRouter()
	router.Use(middlewareHandler.Cors)
	router.Use(middlewareHandler.Panic)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	profileRouter := apiRouter.PathPrefix("/profile").Subrouter()
	gameRouter := apiRouter.PathPrefix("/game").Subrouter()
	adminShopRouter := apiRouter.PathPrefix("/admin").Subrouter()
	shopRouter := apiRouter.PathPrefix("/shop").Subrouter()

	authRouter.HandleFunc("/signin", authHandler.SignIn).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/check", authHandler.Check).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")

	// profileRouter.Use(middlewareHandler.Auth)
	profileRouter.HandleFunc("/{vkid}/rating", profileHandler.GetNearbyUsers).Methods("GET", "OPTIONS")
	profileRouter.HandleFunc("/{vkid}/rating", profileHandler.UpdateRating).Methods("POST", "OPTIONS")
	profileRouter.HandleFunc("/{vkid}/score", profileHandler.GetScore).Methods("GET", "OPTIONS")
	profileRouter.HandleFunc("/{vkid}/superpowers", profileHandler.GetSuperpowers).Methods("GET", "OPTIONS")
	profileRouter.HandleFunc("/{vkid}/superpower/use", profileHandler.UseSuperpower).Methods("POST", "OPTIONS")

	gameRouter.HandleFunc("/rating/top", gameHandler.GetTopUsers).Methods("GET", "OPTIONS")

	// adminShopRouter.Use(middlewareHandler.Auth)
	// adminShopRouter.Use(middlewareHandler.Admin)
	adminShopRouter.HandleFunc("/promocodes", adminShopHandler.GetPromocodes).Methods("GET", "OPTIONS")
	adminShopRouter.HandleFunc("/promocode/add", adminShopHandler.AddPromocode).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/promocode/update", adminShopHandler.UpdatePromocode).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/promocode/delete", adminShopHandler.DeletePromocode).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/products", adminShopHandler.GetProducts).Methods("GET", "OPTIONS")
	adminShopRouter.HandleFunc("/product/add", adminShopHandler.AddProduct).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/product/update", adminShopHandler.UpdateProduct).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/product/delete", adminShopHandler.DeleteProduct).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/tasks", adminShopHandler.GetTasks).Methods("GET", "OPTIONS")
	adminShopRouter.HandleFunc("/task/add", adminShopHandler.AddTask).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/task/update", adminShopHandler.UpdateTask).Methods("POST", "OPTIONS")
	adminShopRouter.HandleFunc("/task/delete", adminShopHandler.DeleteTask).Methods("POST", "OPTIONS")

	apiRouter.HandleFunc("/{vkid}/task", adminShopHandler.PassTask).Methods("POST", "OPTIONS")

	// shopRouter.Use(middlewareHandler.Auth)
	shopRouter.HandleFunc("/tasks", shopHandler.GetTasks).Methods("GET", "OPTIONS")

	return router
}
