package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/router"
	"fmt"
	"log"
)

func main() {
	log.Println("Starting API Gateway...")

	// 1. Load Config
	cfg := config.Load()
	log.Printf("Config loaded: Port=%d, Order=%s, Inventory=%s, Payment=%s",
		cfg.ServerPort, cfg.OrderServiceURL, cfg.InventoryServiceURL, cfg.PaymentServiceURL)

	// 2. Initialize Handlers
	h := handler.NewGatewayHandler(
		cfg.OrderServiceURL,
		cfg.InventoryServiceURL,
		cfg.PaymentServiceURL,
	)
	log.Println("Handlers initialized")

	// 3. Setup Router
	r := router.NewRouter(h)
	log.Println("Router setup complete")

	// 4. Start Server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("API Gateway running on port %d", cfg.ServerPort)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
