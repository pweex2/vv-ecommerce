package main

import (
	"fmt" // 导入 fmt 包用于字符串格式化
	"log"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/router"
	"order-service/internal/service"
	"vv-ecommerce/pkg/clients"
	"vv-ecommerce/pkg/database"

	"gorm.io/driver/mysql" // GORM MySQL 驱动
	"gorm.io/gorm"         // GORM 核心库
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Database connection using GORM
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// TODO: AutoMigrate models if needed
	// db.AutoMigrate(&model.Order{})

	// Initialize repository, service, and handler
	tm := database.NewTransactionManager(db)
	var orderRepo repository.OrderRepository = repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, clients.NewInventoryClient(cfg.InventoryServiceURL), clients.NewPaymentClient(cfg.PaymentServiceURL), tm)
	orderHandler := handler.NewOrderHandler(orderService)

	// Routes
	r := router.NewRouter(orderHandler)

	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Order Service running on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
