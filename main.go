package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"twitter-bookmarks/api"
	"twitter-bookmarks/config"
	"twitter-bookmarks/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	ctx := context.Background()

	twitterService := services.NewTwitterService(cfg.TwitterClientID, cfg.TwitterClientSecret, cfg.TwitterRedirectURI)
	srv := api.New(cfg.Port, api.WithRegisterRoutes(twitterService, cfg.SecretKey, cfg.TwitterAuthToken))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		if err := srv.StartHTTP(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Printf("An error was occurred %v", err)
			}
			<-quit
		}
	}()

	<-quit

	log.Println("shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}

	log.Println("server shutdown")
}
