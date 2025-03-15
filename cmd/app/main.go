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

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/handlers"
	"github.com/gorilla/mux"
)

func main() {
	var (
		host string
		port int
	)

	flag.StringVar(&host, "h", "localhost", "server host")
	flag.IntVar(&port, "p", 8080, "server ip")
	flag.Parse()

	profileHandler, err := handlers.NewProfileHandler()
	if err != nil {
		log.Fatalf("Failed to init profile handler: %v", err)
	}
	gameHandler, err := handlers.NewGameHandler()
	if err != nil {
		log.Fatalf("Failed to init game handler: %v", err)
	}

	router := initRouter(profileHandler, gameHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Failed to stop server gracefully: %v", err)
		}
	}()

	log.Printf("Starting server at http://%s:%d", host, port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}

	<-stopped

	log.Println("Server stopped")
}

func initRouter(profileHandler *handlers.ProfileHandler, gameHandler *handlers.GameHandler) *mux.Router {
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	profileRouter := apiRouter.PathPrefix("/profile").Subrouter()
	gameRouter := apiRouter.PathPrefix("/game").Subrouter()

	profileRouter.HandleFunc("/{uuid}/rating", profileHandler.GetRating).Methods("GET")
	profileRouter.HandleFunc("/{uuid}/rating", profileHandler.UpdateRating).Methods("POST")

	gameRouter.HandleFunc("/rating/top", gameHandler.GetTopRating).Methods("GET")

	return router
}
