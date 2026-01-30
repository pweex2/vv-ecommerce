package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vv-ecommerce/pkg/trace"
)

func main() {
	log.Println("Starting API Gateway...")

	// 1. Load Config
	cfg := config.Load()
	log.Printf("Config loaded: Port=%d, Order=%s, Inventory=%s, Payment=%s, OtelCollectorURL=%s",
		cfg.ServerPort, cfg.OrderServiceURL, cfg.InventoryServiceURL, cfg.PaymentServiceURL, cfg.OtelCollectorURL)

	// 0. Initialize Tracing
	shutdownTracer, err := trace.InitTracer("api-gateway", cfg.OtelCollectorURL)
	if err != nil {
		log.Printf("Failed to init tracer: %v", err)
	}
	defer func() {
		if shutdownTracer != nil {
			if err := shutdownTracer(context.Background()); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}
	}()

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
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("API Gateway running on port %d", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown: signal %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			log.Printf("Could not stop server gracefully: %v", err)
		}
	}
}
