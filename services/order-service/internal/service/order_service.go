package service

import (
	"context"
	"order-service/internal/model"
	"order-service/internal/repository"
	"time"
	"vv-ecommerce/pkg/clients"
	"vv-ecommerce/pkg/common/apperror"
	"vv-ecommerce/pkg/database"

	"github.com/google/uuid"
)

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
	paymentResp, err := s.paymentClient.ProcessPayment(orderID, totalAmount)
	if err != nil {
		// 支付失败，发起库存回滚 (Compensating Transaction)
		s.compensator.Compensate(sku, totalAmount)

		s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusFailed)
		return nil, apperror.Internal("payment processing failed", err)
	}

	if paymentResp.Status != "COMPLETED" {
		// 支付状态非成功，同样需要回滚
		s.compensator.Compensate(sku, totalAmount)

		s.repo.UpdateOrderStatus(ctx, orderID, model.OrderStatusFailed)
		return nil, apperror.Conflict("payment failed with status: "+paymentResp.Status, nil)
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
		return nil, apperror.NotFound("order not found", nil)
	}
	return order, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (int64, error) {
	return s.repo.UpdateOrderStatus(ctx, orderID, status)
}
