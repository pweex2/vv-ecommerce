package service

import (
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
		err = s.inventoryClient.Decrease(sku, reqID, totalAmount)
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
	// assume payment service is completed successfully
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
