package main

import (
	"fmt"
	"log"
	"net/http"

	"inventory-service/internal/config"
	"inventory-service/internal/handler"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	"inventory-service/internal/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Inventory Service is healthy")
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate models
	err = db.AutoMigrate(&model.Inventory{}) // <-- 添加这一行
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	var inventoryRepo repository.InventoryRepository = repository.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/inventories", inventoryHandler.GetInventoriesByProductID) // <-- 添加这一行

	http.HandleFunc("/inventory/sku", inventoryHandler.GetInventoryBySKU)
	http.HandleFunc("/inventory/create", inventoryHandler.CreateInventory)
	http.HandleFunc("/inventory/update", inventoryHandler.UpdateInventory)
	http.HandleFunc("/inventory/decrease", inventoryHandler.DecreaseInventory)

	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Server listening on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
