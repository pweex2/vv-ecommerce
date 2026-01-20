package router

import (
	"fmt"
	"net/http"
	"order-service/internal/handler"
	"vv-ecommerce/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.OrderHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.TraceID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

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
