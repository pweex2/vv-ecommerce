package main

import (
	"fmt" // 导入 fmt 包用于字符串格式化
	"log"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/service"
	"vv-ecommerce/pkg/clients"

	"gorm.io/driver/mysql" // GORM MySQL 驱动
	"gorm.io/gorm"         // GORM 核心库
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Order Service is healthy")
}

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
	var orderRepo repository.OrderRepository = repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, clients.NewInventoryClient(cfg.InventoryServiceURL))
	orderHandler := handler.NewOrderHandler(orderService)

	// Routes
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			orderHandler.CreateOrderHandler(w, r)
		case http.MethodGet:
			orderHandler.GetOrderHandler(w, r)
		case http.MethodPatch:
			orderHandler.UpdateOrderStatusHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/health", healthHandler)

	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Order Service running on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
