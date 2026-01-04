package main

import (
	"fmt"
	"log"
	"net/http"
	"payment-service/internal/config"
	"payment-service/internal/handler"
	"payment-service/internal/model"
	"payment-service/internal/repository"
	"payment-service/internal/router"
	"payment-service/internal/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. ????
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. ?????
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. ??????? (????)
	err = db.AutoMigrate(&model.Payment{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// 4. ?????
	paymentRepo := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 5. Routes
	r := router.NewRouter(paymentHandler)

	// 6. ????
	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Payment Service running on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
