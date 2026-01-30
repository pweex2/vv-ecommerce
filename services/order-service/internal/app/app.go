package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/router"
	"order-service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vv-ecommerce/pkg/async"
	"vv-ecommerce/pkg/clients"
	"vv-ecommerce/pkg/database"
	"vv-ecommerce/pkg/trace"
)

type App struct {
	Cfg             *config.Config
	Router          http.Handler
	OutboxProcessor *service.OutboxProcessor
}

func New(cfg *config.Config) (*App, func(), error) {
	// 0. Initialize Tracing
	shutdownTracer, err := trace.InitTracer("order-service", cfg.OtelCollectorURL)
	if err != nil {
		log.Printf("Failed to init tracer: %v", err)
	}

	// 1. Database
	db, err := database.NewMySQLConnection(database.Config{
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		DBName:   cfg.Database.DBName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 2. Clients
	inventoryClient := clients.NewInventoryClient(cfg.InventoryServiceURL)
	paymentClient := clients.NewPaymentClient(cfg.PaymentServiceURL)

	mqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.MQ.User, cfg.MQ.Password, cfg.MQ.Host, cfg.MQ.Port)
	log.Printf("MQ URL: %s", mqURL)
	messageQueue, err := async.NewRabbitMQ(mqURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 4. Core Logic
	tm := database.NewTransactionManager(db)
	orderRepo := repository.NewOrderRepository(db)
	outboxProcessor := service.NewOutboxProcessor(orderRepo, messageQueue)
	orderService := service.NewOrderService(orderRepo, inventoryClient, paymentClient, tm)
	orderHandler := handler.NewOrderHandler(orderService)

	// 5. Router
	// Note: router package might expose NewRouter or SetupRouter. main.go uses router.NewRouter
	// Checking previous main.go: r := router.NewRouter(orderHandler)
	r := router.NewRouter(orderHandler)

	// Cleanup function
	cleanup := func() {
		log.Println("Cleaning up application resources...")
		outboxProcessor.Stop() // Stop outbox processor
		if err := messageQueue.Close(); err != nil {
			log.Printf("Error closing message queue: %v", err)
		}
		if shutdownTracer != nil {
			if err := shutdownTracer(context.Background()); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}
	}

	return &App{
		Cfg:             cfg,
		Router:          r,
		OutboxProcessor: outboxProcessor,
	}, cleanup, nil
}

func (a *App) Run() error {
	// Start background workers
	a.OutboxProcessor.Start()

	addr := fmt.Sprintf(":%d", a.Cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Order Service running on port %d", a.Cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel to listen for interrupt or terminate signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown: signal %v", sig)

		// Create a context with a timeout for the shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Ask the server to shut down gracefully
		if err := srv.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
