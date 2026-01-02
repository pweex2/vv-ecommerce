package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"inventory-service/internal/service"
	"net/http"
	"strconv"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) GetInventoriesByProductID(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	inventories, err := h.inventoryService.GetInventoriesByProductID(uint(productID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting inventories: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(inventories)
}

func (h *InventoryHandler) GetInventoryBySKU(w http.ResponseWriter, r *http.Request) {
	sku := r.URL.Query().Get("sku")
	if sku == "" {
		http.Error(w, "SKU is required", http.StatusBadRequest)
		return
	}

	inventory, err := h.inventoryService.GetInventoryBySKU(sku)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting inventory: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(inventory)
}

func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID uint   `json:"product_id"`
		SKU       string `json:"sku"`
		Quantity  int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.CreateInventory(req.SKU, req.ProductID, req.Quantity); err != nil {
		http.Error(w, fmt.Sprintf("Error creating inventory: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Inventory added successfully")
}

func (h *InventoryHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SKU      string `json:"sku"`
		Quantity int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.UpdateInventory(req.SKU, req.Quantity); err != nil {
		http.Error(w, fmt.Sprintf("Error updating inventory: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Inventory updated successfully")
}

func (h *InventoryHandler) DecreaseInventory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SKU       string `json:"sku"`
		RequestID string `json:"request_id"`
		Quantity  int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.inventoryService.DecreaseInventory(req.RequestID, req.SKU, req.Quantity); err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDuplicateRequestID) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Error decreasing inventory: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Inventory decreased successfully")
}
