package service

import (
	"context"
	"errors"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/database"

	"gorm.io/gorm"
)

type InventoryService struct {
	repo repository.InventoryRepository
	tm   database.TransactionManager
}

func NewInventoryService(repo repository.InventoryRepository, tm database.TransactionManager) *InventoryService {
	return &InventoryService{repo: repo, tm: tm}
}

func (s *InventoryService) GetInventoriesByProductID(ctx context.Context, productID uint) ([]model.Inventory, error) {
	return s.repo.GetInventoriesByProductID(ctx, productID)
}

func (s *InventoryService) GetInventoryBySKU(ctx context.Context, sku string) (*model.Inventory, error) {
	inventory, err := s.repo.GetInventoryBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("inventory not found", err)
		}
		return nil, apperror.Internal("database error", err)
	}
	return inventory, nil
}

func (s *InventoryService) CreateInventory(ctx context.Context, sku string, productID uint, quantity int) error {
	inventory, err := s.repo.GetInventoryBySKU(ctx, sku)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.Internal("database error", err)
		}
	}
	if inventory != nil {
		return apperror.Conflict("inventory already exists", nil)
	}

	newInventory := &model.Inventory{
		ProductID: productID,
		SKU:       sku,
		Quantity:  quantity,
	}
	if err := s.repo.CreateInventory(ctx, newInventory); err != nil {
		return apperror.Internal("failed to create inventory", err)
	}
	return nil
}

func (s *InventoryService) DecreaseInventory(ctx context.Context, reqID, sku, orderID, traceID string, quantity int) error {
	// 检查是否存在重复请求
	if err := s.repo.RequestLogExists(ctx, reqID); err == nil {
		return apperror.Conflict("duplicate request ID", nil)
	}

	if quantity <= 0 {
		return apperror.InvalidInput("quantity must be positive for decrease operation", nil)
	}

	// 准备日志
	deductionLog := &model.InventoryDeductionLog{
		RequestID: reqID,
		SKU:       sku,
		OrderID:   orderID,
		TraceID:   traceID,
		Quantity:  quantity,
	}

	// 使用原子事务扣减库存并保存日志
	// 业务逻辑控制事务边界：先扣减，再记录日志，两者同生共死
	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		// 1. 扣减库存
		if err := s.repo.DecreaseInventory(ctx, sku, quantity); err != nil {
			return err
		}
		// 2. 记录日志
		if err := s.repo.SaveDeductionLog(ctx, deductionLog); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if err.Error() == "insufficient stock" {
			return apperror.Conflict("insufficient stock", err)
		}
		if err.Error() == "inventory not found" {
			return apperror.NotFound("inventory not found", err)
		}
		return apperror.Internal("transaction failed", err)
	}
	return nil
}

func (s *InventoryService) IncreaseInventory(ctx context.Context, sku string, quantity int, traceID string) error {
	if quantity <= 0 {
		return apperror.InvalidInput("quantity must be positive", nil)
	}
	// TODO: Log traceID for rollback tracking if needed
	// log.Printf("IncreaseInventory (Rollback?) - SKU: %s, Qty: %d, TraceID: %s", sku, quantity, traceID)

	if err := s.repo.IncreaseInventory(ctx, sku, quantity); err != nil {
		return apperror.Internal("failed to increase inventory", err)
	}
	return nil
}

func (s *InventoryService) UpdateInventory(ctx context.Context, sku string, quantity int) error {
	if quantity < 0 {
		return apperror.InvalidInput("quantity cannot be negative", nil)
	}
	inventory, err := s.repo.GetInventoryBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("inventory not found", err)
		}
		return apperror.Internal("database error", err)
	}
	inventory.Quantity = quantity
	if err := s.repo.UpdateInventory(ctx, inventory); err != nil {
		return apperror.Internal("failed to update inventory", err)
	}
	return nil
}
