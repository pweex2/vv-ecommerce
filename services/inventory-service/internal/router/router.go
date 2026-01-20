package router

import (
	"fmt"
	"inventory-service/internal/handler"
	"net/http"
	"vv-ecommerce/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.InventoryHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.TraceID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Inventory Service is healthy")
		fmt.Println("Inventory Service is healthy")
	})

	// Inventory Routes
	r.GET("/inventories", h.GetInventoriesByProductID)
	r.GET("/inventory/sku", h.GetInventoryBySKU)
	r.POST("/inventory/create", h.CreateInventory)
	// r.POST("/inventory/update", h.UpdateInventory) // 暂时移除 UpdateInventory，使用 Increase/Decrease
	r.POST("/inventory/decrease", h.DecreaseInventory)
	r.POST("/inventory/increase", h.IncreaseInventory)

	return r
}
