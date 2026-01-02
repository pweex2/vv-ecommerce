package repository

import (
	"order-service/internal/model"

	"gorm.io/gorm" // 导入 GORM
)

type OrderRepository interface {
	CreateOrder(order *model.Order) error
	GetOrderByID(orderID string) (*model.Order, error)
	UpdateOrderStatus(orderID, status string) (int64, error)
}

type GORMOrderRepository struct {
	db *gorm.DB // 更改为 *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository { // 更改参数类型和返回类型
	return &GORMOrderRepository{db: db}
}

func (r *GORMOrderRepository) CreateOrder(order *model.Order) error {
	return r.db.Create(order).Error // 使用 GORM 的 Create 方法
}

func (r *GORMOrderRepository) GetOrderByID(orderID string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("order_id = ?", orderID).First(&order).Error // 使用 GORM 的 Where 和 First 方法
	if err == gorm.ErrRecordNotFound {
		return nil, nil // Order not found
	}
	return &order, err
}

func (r *GORMOrderRepository) UpdateOrderStatus(orderID, status string) (int64, error) {
	result := r.db.Model(&model.Order{}).Where("order_id = ? AND status != ?", orderID, status).Update("status", status) // 使用 GORM 的 Model, Where 和 Update 方法
	return result.RowsAffected, result.Error
}
