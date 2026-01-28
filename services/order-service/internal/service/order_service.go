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

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, quantity int64, price int64, sku string) (*model.Order, error) {
	orderID := uuid.New().String()
	traceID := uuid.New().String()
	reqID := uuid.New().String()
	var err error

	totalAmount := quantity * price

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
		err = s.inventoryClient.Decrease(sku, reqID, orderID, traceID, quantity)
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
	handleFailure := func(cause error, needRefund bool) error {
		// 1. 如果需要退款 (例如支付成功但后续逻辑失败)，尝试退款
		if needRefund {
			// Best effort refund. If this fails, we need manual intervention or a more robust background job.
			if refundErr := s.paymentClient.Refund(ctx, orderID, traceID); refundErr != nil {
				// Log this critical error. In a real system, send to alert channel.
				// fmt.Printf("CRITICAL: Failed to refund payment for order %s: %v\n", orderID, refundErr)
			}
		}

		// 2. 在一个事务中更新订单状态为失败并记录 Outbox 事件 (回滚库存)
		txErr := s.tm.Transaction(ctx, func(txCtx context.Context) error {
			if _, err := s.repo.UpdateOrderStatus(txCtx, orderID, model.OrderStatusFailed); err != nil {
				return err
			}

			payload := map[string]interface{}{
				"sku":      sku,
				"quantity": quantity,
				"trace_id": traceID,
			}
			outboxEvent := &model.OutboxEvent{
				AggregateType: "Order",
				AggregateID:   orderID,
				EventType:     "InventoryRollback",
				Payload:       datatypes.JSON(mustMarshal(payload)),
				Status:        model.OutboxStatusPending,
				TraceID:       traceID,
			}

			return s.repo.SaveOutboxEvent(txCtx, outboxEvent)
		})

		if txErr != nil {
			return apperror.Internal("payment failed/compensated and persistence failed", txErr)
		}

		return cause
	}

	if err != nil {
		// 支付请求本身失败 (可能是网络错误或 500).
		// 处于不确定状态，为了安全起见，可以尝试退款 (如果对方其实扣款成功了)
		// 但为了简化，这里假设 error 意味着没扣款。
		return nil, handleFailure(apperror.Internal("payment processing failed", err), false)
	}

	if paymentResp.Status != string(constants.PaymentStatusCompleted) {
		return nil, handleFailure(apperror.Conflict("payment failed with status: "+paymentResp.Status, nil), false)
	}

	// 支付成功，进入"危险区"
	// 如果后续步骤失败，必须退款 + 回滚库存

	if _, err := s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusPaid); err != nil {
		return nil, handleFailure(apperror.Internal("failed to update order status to PAID", err), true)
	}

	if _, err := s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusCompleted); err != nil {
		return nil, handleFailure(apperror.Internal("failed to update order status to COMPLETED", err), true)
	}

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
