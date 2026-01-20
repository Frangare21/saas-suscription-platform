package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"saas-subscription-platform/services/api-gateway/internal/config"
	"saas-subscription-platform/services/api-gateway/internal/server"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
	cfg := config.Load()

	srv := server.New(cfg)

	go func() {
		log.Printf("api-gateway running on %s", cfg.HTTPAddr)
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("shutting down api-gateway...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
}
