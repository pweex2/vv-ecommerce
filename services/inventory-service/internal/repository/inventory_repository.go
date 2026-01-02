package repository

import (
	"inventory-service/internal/model"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	GetInventoriesByProductID(productID uint) ([]model.Inventory, error)
	GetInventoryBySKU(sku string) (*model.Inventory, error)
	CreateInventory(inventory *model.Inventory) error
	UpdateInventory(inventory *model.Inventory) error
	RequestLogExists(reqID string) error
	SaveDeductionLog(log *model.InventoryDeductionLog) error
}

type GORMInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &GORMInventoryRepository{db: db}
}

func (r *GORMInventoryRepository) GetInventoriesByProductID(productID uint) ([]model.Inventory, error) {
	var inventories []model.Inventory
	result := r.db.Where("product_id = ?", productID).Find(&inventories)
	if result.Error != nil {
		return nil, result.Error
	}
	return inventories, nil
}

func (r *GORMInventoryRepository) GetInventoryBySKU(sku string) (*model.Inventory, error) {
	var inventory model.Inventory
	result := r.db.Where("sku = ?", sku).First(&inventory)
	if result.Error != nil {
		return nil, result.Error
	}
	return &inventory, nil
}

func (r *GORMInventoryRepository) CreateInventory(inventory *model.Inventory) error {
	return r.db.Create(inventory).Error
}

func (r *GORMInventoryRepository) UpdateInventory(inventory *model.Inventory) error {
	return r.db.Save(inventory).Error
}

func (r *GORMInventoryRepository) RequestLogExists(reqID string) error {
	var log model.InventoryDeductionLog
	result := r.db.Where("request_id = ?", reqID).First(&log)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *GORMInventoryRepository) SaveDeductionLog(log *model.InventoryDeductionLog) error {
	return r.db.Create(log).Error
}
