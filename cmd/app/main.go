package main

import (
	"context"
	"flag"
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

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/handlers"
	storage "github.com/UserNameShouldBeHere/VK-doodle-jump/internal/repository/tarantool"
	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/services"
)

func main() {
	var (
		backendHost          string
		frontendHost         string
		backendPort          int
		frontendtort         int
		leagueUpdateInterval int
	)

	flag.StringVar(&backendHost, "back-h", "127.0.0.1", "backend host")
	flag.StringVar(&frontendHost, "front-h", "127.0.0.1", "frontend host")
	flag.IntVar(&backendPort, "back-p", 80, "backend port")
	flag.IntVar(&frontendtort, "front-p", 3000, "frontend port")
	flag.IntVar(&leagueUpdateInterval, "l-update", 10, "league update interval in seconds")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address:  "tarantool:3301",
		User:     "admin",
		Password: "pass",
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		fmt.Println("Connection refused:", err)
		return
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}
	sugarLogger := logger.Sugar()

	storageCtx, storageCancel := context.WithCancel(context.Background())
	usersStorage, err := storage.NewUsersStorage(storageCtx, conn, leagueUpdateInterval)
	if err != nil {
		log.Fatal(err)
	}
	shopStorage, err := storage.NewShopStorage(conn)
	if err != nil {
		log.Fatal(err)
	}

	usersService, err := services.NewUsersService(usersStorage, sugarLogger)
	if err != nil {
		log.Fatal(err)
	}
	shopService, err := services.NewShopService(shopStorage, sugarLogger)
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
	shopHandler, err := handlers.NewShopHandler(shopService, sugarLogger)
	if err != nil {
		log.Fatalf("Failed to init shop handler: %v", err)
	}
	middlewareHandler, err := handlers.NewMiddlewareHandler(
		fmt.Sprintf("%s:%d", frontendHost, frontendtort),
		logger.Sugar())
	if err != nil {
		log.Fatalf("Failed to init middleware handler: %v", err)
	}

	router := initRouter(gameHandler, profileHandler, shopHandler, middlewareHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", backendPort),
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

	log.Printf("Starting server at http://%s:%d", backendHost, backendPort)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}

	<-stopped

	log.Println("Server stopped")
}

func initRouter(
	gameHandler *handlers.GameHandler,
	profileHandler *handlers.ProfileHandler,
	shopHandler *handlers.ShopHandler,
	middlewareHandler *handlers.MiddlewareHandler) *mux.Router {

	router := mux.NewRouter()
	router.Use(middlewareHandler.Cors)
	router.Use(middlewareHandler.Panic)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	profileRouter := apiRouter.PathPrefix("/profile").Subrouter()
	gameRouter := apiRouter.PathPrefix("/game").Subrouter()
	shopRouter := apiRouter.PathPrefix("/shop").Subrouter()

	profileRouter.HandleFunc("/{uuid}/rating", profileHandler.UpdateRating).Methods("POST", "OPTIONS")

	gameRouter.HandleFunc("/rating/top", gameHandler.GetTopUsers).Methods("GET", "OPTIONS")

	shopRouter.HandleFunc("/promocodes", shopHandler.GetPromocodes).Methods("GET", "OPTIONS")
	shopRouter.HandleFunc("/promocode/add", shopHandler.AddPromocode).Methods("POST", "OPTIONS")
	shopRouter.HandleFunc("/promocode/update", shopHandler.UpdatePromocode).Methods("POST", "OPTIONS")
	shopRouter.HandleFunc("/products", shopHandler.GetProducts).Methods("GET", "OPTIONS")
	shopRouter.HandleFunc("/product/add", shopHandler.AddProduct).Methods("POST", "OPTIONS")
	shopRouter.HandleFunc("/product/update", shopHandler.UpdateProduct).Methods("POST", "OPTIONS")

	return router
}
