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

	"payment-service/internal/config"
	"payment-service/internal/handler"
	"payment-service/internal/model"
	"payment-service/internal/repository"
	"payment-service/internal/router"
	"payment-service/internal/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type App struct {
	Cfg    *config.Config
	Router http.Handler
	DB     *gorm.DB
}

func New(cfg *config.Config) (*App, func(), error) {
	// 1. Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// AutoMigrate models
	err = db.AutoMigrate(&model.Payment{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	// 2. Core Logic
	paymentRepo := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 3. Router
	r := router.NewRouter(paymentHandler)

	// Cleanup function
	cleanup := func() {
		log.Println("Cleaning up application resources...")
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Error getting sql.DB from gorm: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
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
		log.Printf("Payment Service running on port %d", a.Cfg.ServerPort)
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
