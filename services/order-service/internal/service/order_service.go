package service

import (
	"context"
	"order-service/internal/model"
	"order-service/internal/repository"
	"time"
	"vv-ecommerce/pkg/clients"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/common/constants"
	"vv-ecommerce/pkg/database"

	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

type OrderService struct {
	repo            repository.OrderRepository
	inventoryClient *clients.InventoryClient
	paymentClient   *clients.PaymentClient
	compensator     *InventoryCompensator
	tm              database.TransactionManager
}

func NewOrderService(repo repository.OrderRepository, inventoryClient *clients.InventoryClient, paymentClient *clients.PaymentClient, compensator *InventoryCompensator, tm database.TransactionManager) *OrderService {
	return &OrderService{repo: repo, inventoryClient: inventoryClient, paymentClient: paymentClient, compensator: compensator, tm: tm}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, totalAmount int64, sku string) (*model.Order, error) {
	orderID := uuid.New().String()
	traceID := uuid.New().String()
	reqID := uuid.New().String()
	var err error

	order := &model.Order{
		OrderID:     orderID,
		UserID:      userID,
		Status:      model.OrderStatusCreated,
		TotalAmount: totalAmount,
		TraceID:     traceID,
	}

	err = s.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, apperror.Internal("failed to create order", err)
	}

	// retry 3 times
	for i := 0; i < 3; i++ {
		// 调用库存服务减少库存
		err = s.inventoryClient.Decrease(sku, reqID, orderID, traceID, totalAmount)
		if err == nil {
			break
		}
		// Smart Retry: Check if error is retryable
		if !apperror.IsRetryable(err) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusFailed)
		// Inventory client error might be retryable or not, but here we failed after retries
		// 尽量保留原始错误类型，以便上层能区分是 4xx 还是 5xx
		if _, ok := err.(*apperror.AppError); ok {
			return nil, err
		}
		return nil, apperror.Internal("failed to decrease inventory after retries", err)
	}
	s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusInventoryReserved)

	// 调用支付服务创建支付订单
	paymentResp, err := s.paymentClient.ProcessPayment(ctx, orderID, totalAmount, traceID)

	// 定义统一的补偿逻辑
	handlePaymentFailure := func(cause error) error {
		// 在一个事务中更新订单状态为失败并记录 Outbox 事件
		txErr := s.tm.Transaction(ctx, func(txCtx context.Context) error {
			if _, err := s.repo.UpdateOrderStatus(txCtx, orderID, model.OrderStatusFailed); err != nil {
				return err
			}

			payload := map[string]interface{}{
				"sku":      sku,
				"quantity": totalAmount,
				"trace_id": traceID,
			}
			outboxEvent := &model.OutboxEvent{
				AggregateType: "Order",
				AggregateID:   orderID,
				EventType:     "InventoryRollback",
				Payload:       datatypes.JSON(mustMarshal(payload)), // 辅助函数处理 JSON 序列化
				Status:        model.OutboxStatusPending,
				TraceID:       traceID,
			}

			return s.repo.SaveOutboxEvent(txCtx, outboxEvent)
		})

		if txErr != nil {
			// 如果事务提交失败，我们确实处于一个糟糕的状态。
			// 但由于我们还没发 MQ，也没有双写问题。只是 DB 状态更新失败。
			// 日志记录这个严重错误
			// logger.Error("Critical: Failed to save compensation event", txErr)
			return apperror.Internal("payment failed and compensation persistence failed", txErr)
		}

		return cause
	}

	if err != nil {
		return nil, handlePaymentFailure(apperror.Internal("payment processing failed", err))
	}

	if paymentResp.Status != string(constants.PaymentStatusCompleted) {
		return nil, handlePaymentFailure(apperror.Conflict("payment failed with status: "+paymentResp.Status, nil))
	}

	s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusPaid)

	s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusCompleted)

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("database error", err)
	}
	if order == nil {
		return nil, apperror.NotFound("Order not found", nil)
	}
	return order, nil
}

func (s *OrderService) GetOrders(ctx context.Context) ([]*model.Order, error) {
	orders, err := s.repo.GetOrders(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to fetch orders", err)
	}
	// 如果没有订单，返回空切片而不是 nil (虽然 nil slice 序列化也是 null/[]，但显式一点更好)
	if orders == nil {
		orders = []*model.Order{}
	}
	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (int64, error) {
	return s.repo.UpdateOrderStatus(ctx, orderID, status)
}
