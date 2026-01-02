package service

import (
	"errors"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"

	"gorm.io/gorm"
)

var ErrInsufficientStock = errors.New("insufficient stock")
var ErrInventoryNotFound = errors.New("inventory not found")
var ErrDuplicateRequestID = errors.New("duplicate request ID")

type InventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) GetInventoriesByProductID(productID uint) ([]model.Inventory, error) {
	return s.repo.GetInventoriesByProductID(productID)
}

func (s *InventoryService) GetInventoryBySKU(sku string) (*model.Inventory, error) {
	inventory, err := s.repo.GetInventoryBySKU(sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound // 返回更具体的错误
		}
		return nil, err
	}
	return inventory, nil
}

func (s *InventoryService) CreateInventory(sku string, productID uint, quantity int) error {
	inventory, err := s.repo.GetInventoryBySKU(sku)
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
	return s.repo.CreateInventory(newInventory)

}

func (s *InventoryService) DecreaseInventory(reqID, sku string, quantity int) error {
	// 检查是否存在重复请求
	if err := s.repo.RequestLogExists(reqID); err == nil {
		return ErrDuplicateRequestID
	}

	if quantity <= 0 {
		return errors.New("quantity must be positive for decrease operation")
	}
	inventory, err := s.repo.GetInventoryBySKU(sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInventoryNotFound // 如果库存不存在，则无法减少
		}
		return err
	}

	if inventory.Quantity < quantity {
		return ErrInsufficientStock
	}

	inventory.Quantity -= quantity

	err = s.repo.UpdateInventory(inventory)
	if err != nil {
		return err
	}

	// 保存库存减少记录
	deductionLog := &model.InventoryDeductionLog{
		RequestID: reqID,
		SKU:       sku,
		Quantity:  quantity,
	}
	if err := s.repo.SaveDeductionLog(deductionLog); err != nil {
		return err
	}

	return nil
}

func (s *InventoryService) UpdateInventory(sku string, quantity int) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	inventory, err := s.repo.GetInventoryBySKU(sku)
	if err != nil {
		return err
	}
	inventory.Quantity = quantity
	return s.repo.UpdateInventory(inventory)
}
