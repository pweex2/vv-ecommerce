package repository

import (
	"context"
	"order-service/internal/model"
	"vv-ecommerce/pkg/database"

	"gorm.io/gorm" // 导入 GORM
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, orderID string) (*model.Order, error)
	GetOrders(ctx context.Context) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (int64, error)
	SaveOutboxEvent(ctx context.Context, event *model.OutboxEvent) error
	GetPendingOutboxEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	UpdateOutboxEventStatus(ctx context.Context, id uint, status model.OutboxStatus) error
}

type GORMOrderRepository struct {
	db *gorm.DB // 更改为 *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository { // 更改参数类型和返回类型
	return &GORMOrderRepository{db: db}
}

func (r *GORMOrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	return database.GetDB(ctx, r.db).Create(order).Error // 使用 GORM 的 Create 方法
}

func (r *GORMOrderRepository) GetOrderByID(ctx context.Context, orderID string) (*model.Order, error) {
	var order model.Order
	err := database.GetDB(ctx, r.db).Where("order_id = ?", orderID).First(&order).Error // 使用 GORM 的 Where 和 First 方法
	if err == gorm.ErrRecordNotFound {
		return nil, nil // Order not found
	}
	return &order, err
}

func (r *GORMOrderRepository) GetOrders(ctx context.Context) ([]*model.Order, error) {
	var orders []*model.Order
	err := database.GetDB(ctx, r.db).Order("created_at desc").Limit(20).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *GORMOrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (int64, error) {
	result := database.GetDB(ctx, r.db).Model(&model.Order{}).Where("order_id = ? AND status != ?", orderID, status).Update("status", status) // 使用 GORM 的 Model, Where 和 Update 方法
	return result.RowsAffected, result.Error
}

func (r *GORMOrderRepository) SaveOutboxEvent(ctx context.Context, event *model.OutboxEvent) error {
	return database.GetDB(ctx, r.db).Create(event).Error
}

func (r *GORMOrderRepository) GetPendingOutboxEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	err := database.GetDB(ctx, r.db).Where("status = ?", model.OutboxStatusPending).Limit(limit).Order("created_at ASC").Find(&events).Error
	return events, err
}

func (r *GORMOrderRepository) UpdateOutboxEventStatus(ctx context.Context, id uint, status model.OutboxStatus) error {
	return database.GetDB(ctx, r.db).Model(&model.OutboxEvent{}).Where("id = ?", id).Update("status", status).Error
}
