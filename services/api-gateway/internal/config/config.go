package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServerPort          int
	OrderServiceURL     string
	InventoryServiceURL string
	PaymentServiceURL   string
}

func Load() *Config {
	return &Config{
		ServerPort:          getEnvAsInt("SERVER_PORT", 8000), // Gateway 跑在 8000 端口
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "http://localhost:8080"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8081"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	log.Printf("Warning: Invalid integer for env %s: %s. Using default %d", key, strValue, fallback)
	return fallback
}
