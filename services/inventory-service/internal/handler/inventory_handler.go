package handler

import (
	"strconv"

	"inventory-service/internal/service"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/common/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	service *service.InventoryService
}

func NewInventoryHandler(service *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		service: service,
	}
}

func (h *InventoryHandler) GetInventoriesByProductID(c *gin.Context) {
	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		response.Error(c, apperror.InvalidInput("product_id is required", nil))
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		response.Error(c, apperror.InvalidInput("invalid product_id", err))
		return
	}

	inventories, err := h.service.GetInventoriesByProductID(c.Request.Context(), uint(productID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, inventories)
}

func (h *InventoryHandler) GetInventoryBySKU(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		response.Error(c, apperror.InvalidInput("sku is required", nil))
		return
	}

	inventory, err := h.service.GetInventoryBySKU(c.Request.Context(), sku)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, inventory)
}

func (h *InventoryHandler) DecreaseInventory(c *gin.Context) {
	var req struct {
		RequestID string `json:"request_id"` // 可选，如果为空则自动生成
		SKU       string `json:"sku" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,gt=0"`
		OrderID   string `json:"order_id" binding:"required"` // 业务订单号
		TraceID   string `json:"trace_id"`                    // 追踪 ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.InvalidInput("invalid request body or validation failed", err))
		return
	}

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// 修正参数顺序：reqID, sku, orderID, traceID, quantity
	if err := h.service.DecreaseInventory(c.Request.Context(), req.RequestID, req.SKU, req.OrderID, req.TraceID, req.Quantity); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]string{"message": "Inventory decreased successfully"})
}

func (h *InventoryHandler) IncreaseInventory(c *gin.Context) {
	var req struct {
		SKU      string `json:"sku" binding:"required"`
		Quantity int    `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.InvalidInput("invalid request body or validation failed", err))
		return
	}

	if err := h.service.IncreaseInventory(c.Request.Context(), req.SKU, req.Quantity); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]string{"message": "Inventory increased successfully"})
}

func (h *InventoryHandler) CreateInventory(c *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id" binding:"required,gt=0"`
		SKU       string `json:"sku" binding:"required"`
		Quantity  int    `json:"quantity" binding:"gte=0"` // 允许初始为0
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.InvalidInput("invalid request body or validation failed", err))
		return
	}

	if err := h.service.CreateInventory(c.Request.Context(), req.SKU, req.ProductID, req.Quantity); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]string{"message": "Inventory created successfully"})
}

func (h *InventoryHandler) UpdateInventory(c *gin.Context) {
	var req struct {
		SKU      string `json:"sku" binding:"required"`
		Quantity int    `json:"quantity" binding:"gte=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.InvalidInput("invalid request body or validation failed", err))
		return
	}

	if err := h.service.UpdateInventory(c.Request.Context(), req.SKU, req.Quantity); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]string{"message": "Inventory updated successfully"})
}
