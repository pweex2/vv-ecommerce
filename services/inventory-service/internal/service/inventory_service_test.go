package service

import (
	"context"
	"errors"
	"inventory-service/internal/model"
	"testing"
	"vv-ecommerce/pkg/common/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// 1. Mock Repository
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) DecreaseInventory(ctx context.Context, sku string, quantity int) error {
	args := m.Called(ctx, sku, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepository) IncreaseInventory(ctx context.Context, sku string, quantity int) error {
	args := m.Called(ctx, sku, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetInventoryBySKU(ctx context.Context, sku string) (*model.Inventory, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) != nil {
		return args.Get(0).(*model.Inventory), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryRepository) UpdateInventory(ctx context.Context, inventory *model.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetInventoriesByProductID(ctx context.Context, productID uint) ([]model.Inventory, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) != nil {
		return args.Get(0).([]model.Inventory), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryRepository) CreateInventory(ctx context.Context, inventory *model.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

func (m *MockInventoryRepository) RequestLogExists(ctx context.Context, reqID string) error {
	args := m.Called(ctx, reqID)
	return args.Error(0)
}

func (m *MockInventoryRepository) SaveDeductionLog(ctx context.Context, log *model.InventoryDeductionLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetDeductionLog(ctx context.Context, sku, traceID string) (*model.InventoryDeductionLog, error) {
	args := m.Called(ctx, sku, traceID)
	if args.Get(0) != nil {
		return args.Get(0).(*model.InventoryDeductionLog), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryRepository) UpdateDeductionLogStatus(ctx context.Context, id uint, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
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
	// Simulate transaction execution
	return fn(ctx)
}

// 3. Tests
func TestDecreaseInventory_Success(t *testing.T) {
	mockRepo := new(MockInventoryRepository)
	mockTM := new(MockTransactionManager)
	svc := NewInventoryService(mockRepo, mockTM)

	ctx := context.Background()
	reqID := "req-123"
	traceID := "trace-123"
	orderID := "ord-123"
	sku := "SKU-001"
	qty := 5

	// Expectations
	mockRepo.On("RequestLogExists", ctx, reqID).Return(gorm.ErrRecordNotFound) // means not exists
	mockTM.On("Transaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	mockRepo.On("DecreaseInventory", ctx, sku, qty).Return(nil)
	mockRepo.On("SaveDeductionLog", ctx, mock.AnythingOfType("*model.InventoryDeductionLog")).Return(nil)

	// Execute
	err := svc.DecreaseInventory(ctx, reqID, sku, orderID, traceID, qty)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestDecreaseInventory_DuplicateRequest_Fails(t *testing.T) {
	mockRepo := new(MockInventoryRepository)
	mockTM := new(MockTransactionManager)
	svc := NewInventoryService(mockRepo, mockTM)

	ctx := context.Background()
	reqID := "req-duplicate"

	// Expectations
	mockRepo.On("RequestLogExists", ctx, reqID).Return(nil) // nil means it exists!

	// Execute
	err := svc.DecreaseInventory(ctx, reqID, "SKU", "ORD", "TRACE", 5)

	// Assert
	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 40900, appErr.Code) // Conflict
	mockRepo.AssertExpectations(t)
}

func TestRollbackInventory_Success(t *testing.T) {
	mockRepo := new(MockInventoryRepository)
	mockTM := new(MockTransactionManager)
	svc := NewInventoryService(mockRepo, mockTM)

	ctx := context.Background()
	traceID := "trace-rollback"
	sku := "SKU-001"
	qty := 5

	existingLog := &model.InventoryDeductionLog{
		ID:       10,
		SKU:      sku,
		Quantity: qty,
		Status:   "PENDING",
	}

	// Expectations
	mockRepo.On("GetDeductionLog", ctx, sku, traceID).Return(existingLog, nil)
	mockTM.On("Transaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	mockRepo.On("IncreaseInventory", ctx, sku, qty).Return(nil)
	mockRepo.On("UpdateDeductionLogStatus", ctx, uint(10), "ROLLED_BACK").Return(nil)

	// Execute
	err := svc.RollbackInventory(ctx, sku, qty, traceID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRollbackInventory_NoLog_Skips(t *testing.T) {
	mockRepo := new(MockInventoryRepository)
	mockTM := new(MockTransactionManager)
	svc := NewInventoryService(mockRepo, mockTM)

	ctx := context.Background()
	traceID := "trace-nolog"
	sku := "SKU-001"

	// Expectations
	mockRepo.On("GetDeductionLog", ctx, sku, traceID).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	err := svc.RollbackInventory(ctx, sku, 5, traceID)

	// Assert
	assert.NoError(t, err) // Should return nil for idempotency
	mockRepo.AssertExpectations(t)
	mockTM.AssertNotCalled(t, "Transaction", mock.Anything, mock.Anything)
}
