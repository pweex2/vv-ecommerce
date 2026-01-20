package app

import (
	"fmt"
	"log"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/router"
	"order-service/internal/service"
	"vv-ecommerce/pkg/async"
	"vv-ecommerce/pkg/clients"
	"vv-ecommerce/pkg/database"
)

type App struct {
	Cfg         *config.Config
	Router      http.Handler
	Compensator *service.InventoryCompensator
}

func New(cfg *config.Config) (*App, func(), error) {
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

	// 3. MQ
	mqUser := cfg.MQ.User
	if mqUser == "" {
		mqUser = "guest"
	}
	mqPass := cfg.MQ.Password
	if mqPass == "" {
		mqPass = "guest"
	}

	mqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", mqUser, mqPass, cfg.MQ.Host, cfg.MQ.Port)
	messageQueue := async.NewRabbitMQOrMemory(mqURL)

	// 4. Core Logic
	tm := database.NewTransactionManager(db)
	orderRepo := repository.NewOrderRepository(db)
	compensator := service.NewInventoryCompensator(inventoryClient, messageQueue)
	orderService := service.NewOrderService(orderRepo, inventoryClient, paymentClient, compensator, tm)
	orderHandler := handler.NewOrderHandler(orderService)

	// 5. Router
	// Note: router package might expose NewRouter or SetupRouter. main.go uses router.NewRouter
	// Checking previous main.go: r := router.NewRouter(orderHandler)
	r := router.NewRouter(orderHandler)

	// Cleanup function
	cleanup := func() {
		log.Println("Cleaning up application resources...")
		if err := messageQueue.Close(); err != nil {
			log.Printf("Error closing message queue: %v", err)
		}
	}

	return &App{
		Cfg:         cfg,
		Router:      r,
		Compensator: compensator,
	}, cleanup, nil
}

func (a *App) Run() error {
	// Start background workers
	a.Compensator.StartWorker()

	addr := fmt.Sprintf(":%d", a.Cfg.ServerPort)
	log.Printf("Order Service running on port %d", a.Cfg.ServerPort)
	return http.ListenAndServe(addr, a.Router)
}
