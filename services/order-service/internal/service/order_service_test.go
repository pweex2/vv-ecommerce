package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"order-service/internal/model"
	"testing"
	"vv-ecommerce/pkg/clients"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Mock Repository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, orderID string) (*model.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*model.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) GetOrders(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]*model.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status model.OrderStatus) (int64, error) {
	args := m.Called(ctx, orderID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) SaveOutboxEvent(ctx context.Context, event *model.OutboxEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOrderRepository) GetPendingOutboxEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]model.OutboxEvent), args.Error(1)
}

func (m *MockOrderRepository) UpdateOutboxEventStatus(ctx context.Context, id uint, status model.OutboxStatus, lastError string) error {
	args := m.Called(ctx, id, status, lastError)
	return args.Error(0)
}

func (m *MockOrderRepository) IncrementRetryCount(ctx context.Context, id uint, lastError string) error {
	args := m.Called(ctx, id, lastError)
	return args.Error(0)
}

func (m *MockOrderRepository) FetchAndLockPendingEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]model.OutboxEvent), args.Error(1)
}

// 2. Mock Transaction Manager
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	return fn(ctx)
}

// 3. Tests
func TestCreateOrder_Success(t *testing.T) {
	// Setup Mock HTTP Servers for downstream dependencies
	invServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/inventory/decrease", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer invServer.Close()

	payServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/payments", r.URL.Path)
		resp := clients.PaymentResponse{Status: "COMPLETED"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer payServer.Close()

	// Initialize dependencies
	invClient := clients.NewInventoryClient(invServer.URL, "")
	payClient := clients.NewPaymentClient(payServer.URL)
	mockRepo := new(MockOrderRepository)
	mockTM := new(MockTransactionManager)
	svc := NewOrderService(mockRepo, invClient, payClient, mockTM)

	ctx := context.Background()

	// Expectations
	mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("*model.Order")).Return(nil)
	mockRepo.On("UpdateOrderStatus", ctx, mock.AnythingOfType("string"), model.OrderStatusInventoryReserved).Return(int64(1), nil)
	mockRepo.On("UpdateOrderStatus", ctx, mock.AnythingOfType("string"), model.OrderStatusPaid).Return(int64(1), nil)
	mockRepo.On("UpdateOrderStatus", ctx, mock.AnythingOfType("string"), model.OrderStatusCompleted).Return(int64(1), nil)

	// Execute
	order, err := svc.CreateOrder(ctx, 1, 2, 100, "SKU-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(200), order.TotalAmount)
	mockRepo.AssertExpectations(t)
}

func TestCreateOrder_InventoryFails(t *testing.T) {
	invServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Simulate insufficient stock
		w.Write([]byte(`{"code": 409, "msg": "insufficient stock"}`))
	}))
	defer invServer.Close()

	payServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Should not be called
	}))
	defer payServer.Close()

	invClient := clients.NewInventoryClient(invServer.URL, "")
	payClient := clients.NewPaymentClient(payServer.URL)
	mockRepo := new(MockOrderRepository)
	mockTM := new(MockTransactionManager)
	svc := NewOrderService(mockRepo, invClient, payClient, mockTM)

	ctx := context.Background()

	// Expectations
	mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("*model.Order")).Return(nil)
	mockRepo.On("UpdateOrderStatus", ctx, mock.AnythingOfType("string"), model.OrderStatusFailed).Return(int64(1), nil)

	// Execute
	order, err := svc.CreateOrder(ctx, 1, 2, 100, "SKU-123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	mockRepo.AssertExpectations(t)
}
