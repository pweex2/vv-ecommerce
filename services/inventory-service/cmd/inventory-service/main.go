package main

import (
	"fmt"
	"log"
	"net/http"

	"inventory-service/internal/config"
	"inventory-service/internal/handler"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	"inventory-service/internal/router"
	"inventory-service/internal/service"
	"vv-ecommerce/pkg/database"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

	tm := database.NewTransactionManager(db)
	var inventoryRepo repository.InventoryRepository = repository.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo, tm)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	r := router.NewRouter(inventoryHandler)

	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Server listening on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
