package handler

import (
	"fmt"
	"net/http"
	"order-service/internal/model"
	"order-service/internal/service"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/common/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: s,
	}
}

func (h *OrderHandler) CreateOrderHandler(c *gin.Context) {
	var input struct {
		UserID   int64  `json:"user_id" binding:"required,gt=0"`
		Quantity int64  `json:"quantity" binding:"required,gt=0"`
		Price    int64  `json:"price" binding:"required,gt=0"`
		SKU      string `json:"sku" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.InvalidInput("invalid request body or validation failed", err))
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), input.UserID, input.Quantity, input.Price, input.SKU)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrderHandler(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		response.Error(c, apperror.InvalidInput("Missing order_id", nil))
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if order == nil {
		response.Error(c, apperror.NotFound("Order not found", nil))
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) ListOrdersHandler(c *gin.Context) {
	orders, err := h.service.GetOrders(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) UpdateOrderStatusHandler(c *gin.Context) {
	var input struct {
		OrderID string            `json:"order_id" binding:"required"`
		Status  model.OrderStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.InvalidInput("invalid request body or validation failed", err))
		return
	}

	rowsAffected, err := h.service.UpdateOrderStatus(c.Request.Context(), input.OrderID, input.Status)
	if err != nil {
		response.Error(c, err)
		return
	}

	if rowsAffected == 0 {
		response.Error(c, apperror.NotFound("No update needed or order not found", nil))
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Order %s updated to %s", input.OrderID, input.Status)})
}
