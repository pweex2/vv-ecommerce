package repository

import (
	"context"
	"errors"
	"inventory-service/internal/model"
	"vv-ecommerce/pkg/database"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	GetInventoriesByProductID(ctx context.Context, productID uint) ([]model.Inventory, error)
	GetInventoryBySKU(ctx context.Context, sku string) (*model.Inventory, error)
	CreateInventory(ctx context.Context, inventory *model.Inventory) error
	UpdateInventory(ctx context.Context, inventory *model.Inventory) error
	RequestLogExists(ctx context.Context, reqID string) error
	SaveDeductionLog(ctx context.Context, log *model.InventoryDeductionLog) error
	DecreaseInventory(ctx context.Context, sku string, quantity int) error
	IncreaseInventory(ctx context.Context, sku string, quantity int) error
}

type GORMInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &GORMInventoryRepository{db: db}
}

func (r *GORMInventoryRepository) GetInventoriesByProductID(ctx context.Context, productID uint) ([]model.Inventory, error) {
	var inventories []model.Inventory
	result := database.GetDB(ctx, r.db).Where("product_id = ?", productID).Find(&inventories)
	if result.Error != nil {
		return nil, result.Error
	}
	return inventories, nil
}

func (r *GORMInventoryRepository) GetInventoryBySKU(ctx context.Context, sku string) (*model.Inventory, error) {
	var inventory model.Inventory
	result := database.GetDB(ctx, r.db).Where("sku = ?", sku).First(&inventory)
	if result.Error != nil {
		return nil, result.Error
	}
	return &inventory, nil
}

func (r *GORMInventoryRepository) CreateInventory(ctx context.Context, inventory *model.Inventory) error {
	return database.GetDB(ctx, r.db).Create(inventory).Error
}

func (r *GORMInventoryRepository) UpdateInventory(ctx context.Context, inventory *model.Inventory) error {
	return database.GetDB(ctx, r.db).Save(inventory).Error
}

func (r *GORMInventoryRepository) RequestLogExists(ctx context.Context, reqID string) error {
	var log model.InventoryDeductionLog
	result := database.GetDB(ctx, r.db).Where("request_id = ?", reqID).First(&log)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *GORMInventoryRepository) SaveDeductionLog(ctx context.Context, log *model.InventoryDeductionLog) error {
	return database.GetDB(ctx, r.db).Create(log).Error
}

func (r *GORMInventoryRepository) DecreaseInventory(ctx context.Context, sku string, quantity int) error {
	// 1. Check stock and decrease atomically
	// UPDATE inventories SET quantity = quantity - ? WHERE sku = ? AND quantity >= ?
	result := database.GetDB(ctx, r.db).Model(&model.Inventory{}).
		Where("sku = ? AND quantity >= ?", sku, quantity).
		Update("quantity", gorm.Expr("quantity - ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Either SKU not found or insufficient stock
		// Let's check which one it is to return a better error
		var count int64
		database.GetDB(ctx, r.db).Model(&model.Inventory{}).Where("sku = ?", sku).Count(&count)
		if count == 0 {
			return errors.New("inventory not found")
		}
		return errors.New("insufficient stock")
	}

	return nil
}

func (r *GORMInventoryRepository) IncreaseInventory(ctx context.Context, sku string, quantity int) error {
	result := database.GetDB(ctx, r.db).Model(&model.Inventory{}).
		Where("sku = ?", sku).
		Update("quantity", gorm.Expr("quantity + ?", quantity))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("inventory not found")
	}
	return nil
}
