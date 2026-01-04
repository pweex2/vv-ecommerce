package handler

import (
	"fmt"
	"net/http"
	"order-service/internal/model"
	"order-service/internal/service"
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
		UserID      int64  `json:"user_id" binding:"required,gt=0"`
		TotalAmount int64  `json:"total_amount" binding:"required,gt=0"`
		SKU         string `json:"sku" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body or validation failed")
		return
	}

	order, err := h.service.CreateOrder(input.UserID, input.TotalAmount, input.SKU)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) GetOrderHandler(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		response.Error(c, http.StatusBadRequest, "Missing order_id")
		return
	}

	order, err := h.service.GetOrder(orderID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Order not found")
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) UpdateOrderStatusHandler(c *gin.Context) {
	var input struct {
		OrderID string            `json:"order_id" binding:"required"`
		Status  model.OrderStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body or validation failed")
		return
	}

	rowsAffected, err := h.service.UpdateOrderStatus(input.OrderID, input.Status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if rowsAffected == 0 {
		response.Error(c, http.StatusBadRequest, "No update needed or order not found")
		return
	}

	response.Success(c, map[string]string{"message": fmt.Sprintf("Order %s updated to %s", input.OrderID, input.Status)})
}
