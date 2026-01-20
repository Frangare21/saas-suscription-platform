package main

import (
	"log"
	"saas-subscription-platform/services/billing-service/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
