package handler

import (
	"net/http"
	"payment-service/internal/service"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/common/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(s *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: s,
	}
}

type ProcessPaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required,gt=0"`
}

func (h *PaymentHandler) ProcessPaymentHandler(c *gin.Context) {
	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.InvalidInput("Invalid request body or validation failed", err))
		return
	}

	payment, err := h.service.ProcessPayment(req.OrderID, req.Amount)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetPaymentHandler(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		response.Error(c, apperror.InvalidInput("Missing order_id", nil))
		return
	}

	payment, err := h.service.GetPayment(orderID)
	if err != nil {
		response.Error(c, apperror.NotFound("Payment not found", err))
		return
	}

	c.JSON(http.StatusOK, payment)
}
