package main

import (
	"log"

	"payment-service/internal/app"
	"payment-service/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	application, cleanup, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer cleanup()

	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
