package router

import (
	"fmt"
	"net/http"
	"payment-service/internal/handler"
	"vv-ecommerce/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.PaymentHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.TraceID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Payment Service is healthy")
		fmt.Println("Payment Service is healthy")
	})

	// Payment Routes
	r.POST("/payments", h.ProcessPaymentHandler)
	r.GET("/payments", h.GetPaymentHandler)

	return r
}
