package router

import (
	"fmt"
	"net/http"
	"order-service/internal/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.OrderHandler) *gin.Engine {
	r := gin.Default()

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Order Service is healthy")
		fmt.Println("Order Service is healthy")
	})

	// Order Routes
	r.POST("/orders", h.CreateOrderHandler)
	r.GET("/orders", h.GetOrderHandler)
	r.PATCH("/orders", h.UpdateOrderStatusHandler)

	return r
}
