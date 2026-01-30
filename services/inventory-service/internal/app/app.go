package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-service/internal/config"
	"inventory-service/internal/handler"
	"inventory-service/internal/repository"
	"inventory-service/internal/router"
	"inventory-service/internal/service"
	"vv-ecommerce/pkg/async"
	"vv-ecommerce/pkg/database"
	"vv-ecommerce/pkg/trace"

	"gorm.io/gorm"
)

type App struct {
	Cfg    *config.Config
	Router http.Handler
	DB     *gorm.DB
}

func New(cfg *config.Config) (*App, func(), error) {
	// 0. Initialize Tracing
	shutdownTracer, err := trace.InitTracer("inventory-service", cfg.OtelCollectorURL)
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

	// 2. Core Logic
	tm := database.NewTransactionManager(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo, tm)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	// 2.5 Async Messaging (RabbitMQ)
	mqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.MQ.User, cfg.MQ.Password, cfg.MQ.Host, cfg.MQ.Port)
	mq, err := async.NewRabbitMQ(mqURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Initialize Async Handler
	eventHandler := handler.NewEventHandler(inventoryService, mq)
	if err := eventHandler.RegisterSubscribers(); err != nil {
		return nil, nil, fmt.Errorf("failed to register subscribers: %w", err)
	}

	// 3. Router
	r := router.NewRouter(inventoryHandler)

	// Cleanup function
	cleanup := func() {
		log.Println("Cleaning up application resources...")
		if mq != nil {
			mq.Close()
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Error getting sql.DB from gorm: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
		if shutdownTracer != nil {
			if err := shutdownTracer(context.Background()); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}
	}

	return &App{
		Cfg:    cfg,
		Router: r,
		DB:     db,
	}, cleanup, nil
}

func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.Cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Inventory Service running on port %d", a.Cfg.ServerPort)
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
