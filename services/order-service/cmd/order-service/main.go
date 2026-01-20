package main

import (
	"log"
	"order-service/internal/app"
	"order-service/internal/config"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Application
	application, cleanup, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer cleanup()

	// 3. Run Application
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed to run: %v", err)
	}
}
