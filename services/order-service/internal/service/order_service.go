package service

import (
	"errors"
	"order-service/internal/model"
	"order-service/internal/repository"
	"time"
	"vv-ecommerce/pkg/clients"

	"github.com/google/uuid"
)

type OrderService struct {
	repo            repository.OrderRepository
	inventoryClient *clients.InventoryClient
	paymentClient   *clients.PaymentClient
}

func NewOrderService(repo repository.OrderRepository, inventoryClient *clients.InventoryClient, paymentClient *clients.PaymentClient) *OrderService {
	return &OrderService{repo: repo, inventoryClient: inventoryClient, paymentClient: paymentClient}
}

func (s *OrderService) CreateOrder(userID int64, totalAmount int64, sku string) (*model.Order, error) {
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

	err = s.repo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	// retry 3 times
	for i := 0; i < 3; i++ {
		// 调用库存服务减少库存
		err = s.inventoryClient.Decrease(sku, reqID, orderID, traceID, totalAmount)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		s.repo.UpdateOrderStatus(orderID, model.OrderStatusFailed)
		return nil, err
	}
	s.repo.UpdateOrderStatus(orderID, model.OrderStatusInventoryReserved)

	// 调用支付服务创建支付订单
	paymentResp, err := s.paymentClient.ProcessPayment(orderID, totalAmount)
	if err != nil {
		// 支付失败，发起库存回滚 (Compensating Transaction)
		// 注意：实际生产中，回滚操作也可能失败，因此通常需要将回滚任务放入消息队列 (MQ) 进行异步重试
		// 这里为了演示简单，直接同步调用，仅记录错误日志
		if rollbackErr := s.inventoryClient.Increase(sku, totalAmount); rollbackErr != nil {
			// Log this critical error! "Failed to rollback inventory for order %s: %v"
			// In a real system, alert on call
		}

		s.repo.UpdateOrderStatus(orderID, model.OrderStatusFailed)
		return nil, err
	}

	if paymentResp.Status != "COMPLETED" {
		// 支付状态非成功，同样需要回滚
		if rollbackErr := s.inventoryClient.Increase(sku, totalAmount); rollbackErr != nil {
			// Log critical error
		}

		s.repo.UpdateOrderStatus(orderID, model.OrderStatusFailed)
		return nil, errors.New("payment failed with status: " + paymentResp.Status)
	}

	s.repo.UpdateOrderStatus(orderID, model.OrderStatusPaid)

	s.repo.UpdateOrderStatus(orderID, model.OrderStatusCompleted)

	return order, nil
}

func (s *OrderService) GetOrder(orderID string) (*model.Order, error) {
	return s.repo.GetOrderByID(orderID)
}

func (s *OrderService) UpdateOrderStatus(orderID string, status model.OrderStatus) (int64, error) {
	return s.repo.UpdateOrderStatus(orderID, status)
}
