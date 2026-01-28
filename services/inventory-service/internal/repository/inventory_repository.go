package repository

import (
	"context"
	"inventory-service/internal/model"
	"vv-ecommerce/pkg/database"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	DecreaseInventory(ctx context.Context, sku string, quantity int) error
	IncreaseInventory(ctx context.Context, sku string, quantity int) error
	GetInventoryBySKU(ctx context.Context, sku string) (*model.Inventory, error)
	UpdateInventory(ctx context.Context, inventory *model.Inventory) error
	GetInventoriesByProductID(ctx context.Context, productID uint) ([]model.Inventory, error)
	CreateInventory(ctx context.Context, inventory *model.Inventory) error
	RequestLogExists(ctx context.Context, reqID string) error
	SaveDeductionLog(ctx context.Context, log *model.InventoryDeductionLog) error
	GetDeductionLog(ctx context.Context, sku, traceID string) (*model.InventoryDeductionLog, error)
	UpdateDeductionLogStatus(ctx context.Context, id uint, status string) error
}

type GORMInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &GORMInventoryRepository{db: db}
}

func (r *GORMInventoryRepository) DecreaseInventory(ctx context.Context, sku string, quantity int) error {
	result := database.GetDB(ctx, r.db).Model(&model.Inventory{}).
		Where("sku = ? AND quantity >= ?", sku, quantity).
		Update("quantity", gorm.Expr("quantity - ?", quantity))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Or a custom "insufficient stock" error if you check existence first
	}
	return nil
}

func (r *GORMInventoryRepository) IncreaseInventory(ctx context.Context, sku string, quantity int) error {
	return database.GetDB(ctx, r.db).Model(&model.Inventory{}).
		Where("sku = ?", sku).
		Update("quantity", gorm.Expr("quantity + ?", quantity)).Error
}

func (r *GORMInventoryRepository) GetInventoryBySKU(ctx context.Context, sku string) (*model.Inventory, error) {
	var inventory model.Inventory
	if err := database.GetDB(ctx, r.db).Where("sku = ?", sku).First(&inventory).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *GORMInventoryRepository) UpdateInventory(ctx context.Context, inventory *model.Inventory) error {
	return database.GetDB(ctx, r.db).Save(inventory).Error
}

func (r *GORMInventoryRepository) GetInventoriesByProductID(ctx context.Context, productID uint) ([]model.Inventory, error) {
	var inventories []model.Inventory
	if err := database.GetDB(ctx, r.db).Where("product_id = ?", productID).Find(&inventories).Error; err != nil {
		return nil, err
	}
	return inventories, nil
}

func (r *GORMInventoryRepository) CreateInventory(ctx context.Context, inventory *model.Inventory) error {
	return database.GetDB(ctx, r.db).Create(inventory).Error
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

func (r *GORMInventoryRepository) GetDeductionLog(ctx context.Context, sku, traceID string) (*model.InventoryDeductionLog, error) {
	var log model.InventoryDeductionLog
	// Try to find by TraceID and SKU (since rollback message contains these)
	// Ideally we should use RequestID but OrderService rollback message doesn't carry it yet.
	// Assuming 1 deduction per SKU per TraceID for now.
	if err := database.GetDB(ctx, r.db).Where("trace_id = ? AND sku = ?", traceID, sku).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *GORMInventoryRepository) UpdateDeductionLogStatus(ctx context.Context, id uint, status string) error {
	return database.GetDB(ctx, r.db).Model(&model.InventoryDeductionLog{}).Where("id = ?", id).Update("status", status).Error
}
