package service

import (
	"context"
	"errors"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	"vv-ecommerce/pkg/database"

	"gorm.io/gorm"
)

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrInventoryNotFound = errors.New("inventory not found")
var ErrDuplicateRequestID = errors.New("duplicate request ID")

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
			return nil, ErrInventoryNotFound // 返回更具体的错误
		}
		return nil, err
	}
	return inventory, nil
}

func (s *InventoryService) CreateInventory(ctx context.Context, sku string, productID uint, quantity int) error {
	inventory, err := s.repo.GetInventoryBySKU(ctx, sku)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	if inventory != nil {
		return errors.New("inventory already exists")
	}

	newInventory := &model.Inventory{
		ProductID: productID,
		SKU:       sku,
		Quantity:  quantity,
	}
	return s.repo.CreateInventory(ctx, newInventory)

}

func (s *InventoryService) DecreaseInventory(ctx context.Context, reqID, sku, orderID, traceID string, quantity int) error {
	// 检查是否存在重复请求
	if err := s.repo.RequestLogExists(ctx, reqID); err == nil {
		return ErrDuplicateRequestID
	}

	if quantity <= 0 {
		return errors.New("quantity must be positive for decrease operation")
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
			return ErrInsufficientStock
		}
		if err.Error() == "inventory not found" {
			return ErrInventoryNotFound
		}
		return err
	}

	return nil
}

func (s *InventoryService) IncreaseInventory(ctx context.Context, sku string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	return s.repo.IncreaseInventory(ctx, sku, quantity)
}

func (s *InventoryService) UpdateInventory(ctx context.Context, sku string, quantity int) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	inventory, err := s.repo.GetInventoryBySKU(ctx, sku)
	if err != nil {
		return err
	}
	inventory.Quantity = quantity
	return s.repo.UpdateInventory(ctx, inventory)
}
