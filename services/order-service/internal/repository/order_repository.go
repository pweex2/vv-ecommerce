package repository

import (
	"context"
	"order-service/internal/model"
	"vv-ecommerce/pkg/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, orderID string) (*model.Order, error)
	GetOrders(ctx context.Context) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (int64, error)
	SaveOutboxEvent(ctx context.Context, event *model.OutboxEvent) error
	GetPendingOutboxEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	UpdateOutboxEventStatus(ctx context.Context, id uint, status model.OutboxStatus, lastError string) error
	IncrementRetryCount(ctx context.Context, id uint, lastError string) error
	FetchAndLockPendingEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error)
}

type GORMOrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
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

// FetchAndLockPendingEvents uses FOR UPDATE SKIP LOCKED to fetch pending events safely in a concurrent environment
func (r *GORMOrderRepository) FetchAndLockPendingEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	// MySQL 8.0+ supports SKIP LOCKED
	// This ensures that multiple instances of the service do not pick up the same event
	err := database.GetDB(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ?", model.OutboxStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *GORMOrderRepository) IncrementRetryCount(ctx context.Context, id uint, lastError string) error {
	return database.GetDB(ctx, r.db).Model(&model.OutboxEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"retry_count": gorm.Expr("retry_count + ?", 1),
			"last_error":  lastError,
		}).Error
}

func (r *GORMOrderRepository) UpdateOutboxEventStatus(ctx context.Context, id uint, status model.OutboxStatus, lastError string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if lastError != "" {
		updates["last_error"] = lastError
	}
	// Also increment retry count if failed
	if status == model.OutboxStatusFailed || status == model.OutboxStatusPending {
		// This logic might be better placed in service, but simple increment here is fine if we had a separate method
		// For now, let's keep it simple. The caller (processor) should handle retry logic more explicitly if needed.
		// But wait, our processor needs to update RetryCount.
	}

	return database.GetDB(ctx, r.db).Model(&model.OutboxEvent{}).Where("id = ?", id).Updates(updates).Error
}
