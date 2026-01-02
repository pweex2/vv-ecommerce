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
}

func NewOrderService(repo repository.OrderRepository, inventoryClient *clients.InventoryClient) *OrderService {
	return &OrderService{repo: repo, inventoryClient: inventoryClient}
}

func (s *OrderService) CreateOrder(userID int64, totalAmount int64) (*model.Order, error) {
	orderID := uuid.New().String()
	traceID := uuid.New().String()
	reqID := uuid.New().String()
	var err error

	// retry 3 times
	for i := 0; i < 3; i++ {
		// 调用库存服务减少库存
		err = s.inventoryClient.Decrease("aimu-sohai-red", reqID, totalAmount)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}

	order := &model.Order{
		OrderID:     orderID,
		UserID:      userID,
		Status:      "INIT",
		TotalAmount: totalAmount,
		TraceID:     traceID,
	}

	err = s.repo.CreateOrder(order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetOrder(orderID string) (*model.Order, error) {
	return s.repo.GetOrderByID(orderID)
}

func (s *OrderService) UpdateOrderStatus(orderID, status string) (int64, error) {
	return s.repo.UpdateOrderStatus(orderID, status)
}
