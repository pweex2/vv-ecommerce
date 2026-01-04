package handler

import (
	"net/http"
	"strconv"

	"inventory-service/internal/service"
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
		response.Error(c, http.StatusBadRequest, "product_id is required")
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product_id")
		return
	}

	inventories, err := h.service.GetInventoriesByProductID(uint(productID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, inventories)
}

func (h *InventoryHandler) GetInventoryBySKU(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		response.Error(c, http.StatusBadRequest, "sku is required")
		return
	}

	inventory, err := h.service.GetInventoryBySKU(sku)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, inventory)
}

func (h *InventoryHandler) DecreaseInventory(c *gin.Context) {
	var req struct {
		RequestID string `json:"request_id"` // 可选，如果为空则自动生成
		SKU       string `json:"sku" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body or validation failed")
		return
	}

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// 修正参数顺序：reqID, sku, quantity
	if err := h.service.DecreaseInventory(req.RequestID, req.SKU, req.Quantity); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Inventory decreased successfully"})
}

func (h *InventoryHandler) CreateInventory(c *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id" binding:"required,gt=0"`
		SKU       string `json:"sku" binding:"required"`
		Quantity  int    `json:"quantity" binding:"gte=0"` // 允许初始为0
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body or validation failed")
		return
	}

	if err := h.service.CreateInventory(req.SKU, req.ProductID, req.Quantity); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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
		response.Error(c, http.StatusBadRequest, "invalid request body or validation failed")
		return
	}

	if err := h.service.UpdateInventory(req.SKU, req.Quantity); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Inventory updated successfully"})
}
