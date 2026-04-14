package service

import (
	"errors"
	"payment-service/internal/model"
	"testing"
	"vv-ecommerce/pkg/common/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Define the Mock Repository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePayment(payment *model.Payment) error {
	args := m.Called(payment)
	// We might mutate the payment to simulate DB behavior (e.g. setting ID)
	if args.Get(0) != nil {
		payment.ID = 1
	}
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPaymentByOrderID(orderID string) (*model.Payment, error) {
	args := m.Called(orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*model.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) UpdatePaymentStatus(paymentID uint, status string, transactionID string) error {
	args := m.Called(paymentID, status, transactionID)
	return args.Error(0)
}

// 2. Unit Test for ProcessPayment
func TestProcessPayment_Success(t *testing.T) {
	// Setup Mock
	mockRepo := new(MockPaymentRepository)
	svc := NewPaymentService(mockRepo)

	orderID := "ORD-12345"
	amount := int64(1000)

	// Expectations
	// Expect CreatePayment to be called once
	mockRepo.On("CreatePayment", mock.AnythingOfType("*model.Payment")).Return(nil)

	// Expect UpdatePaymentStatus to be called once with COMPLETED status
	mockRepo.On("UpdatePaymentStatus", uint(0), string(constants.PaymentStatusCompleted), mock.AnythingOfType("string")).Return(nil)

	// Execute
	payment, err := svc.ProcessPayment(orderID, amount)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, string(constants.PaymentStatusCompleted), payment.Status)
	assert.NotEmpty(t, payment.TransactionID)

	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
}

func TestProcessPayment_NegativeAmount_Fails(t *testing.T) {
	// Setup Mock
	mockRepo := new(MockPaymentRepository)
	svc := NewPaymentService(mockRepo)

	orderID := "ORD-INVALID"
	amount := int64(-500)

	// Expectations
	mockRepo.On("CreatePayment", mock.AnythingOfType("*model.Payment")).Return(nil)
	mockRepo.On("UpdatePaymentStatus", uint(0), string(constants.PaymentStatusFailed), "").Return(nil)

	// Execute
	payment, err := svc.ProcessPayment(orderID, amount)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "invalid amount", err.Error())
	assert.Equal(t, string(constants.PaymentStatusFailed), payment.Status)

	mockRepo.AssertExpectations(t)
}

// 3. Unit Test for RefundPayment
func TestRefundPayment_Success(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	svc := NewPaymentService(mockRepo)

	orderID := "ORD-REFUND-1"

	existingPayment := &model.Payment{
		ID:      1,
		OrderID: orderID,
		Status:  string(constants.PaymentStatusCompleted),
	}

	// Expectations
	mockRepo.On("GetPaymentByOrderID", orderID).Return(existingPayment, nil)
	mockRepo.On("UpdatePaymentStatus", uint(1), string(constants.PaymentStatusRefunded), mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := svc.RefundPayment(orderID)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRefundPayment_NotCompleted_Fails(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	svc := NewPaymentService(mockRepo)

	orderID := "ORD-REFUND-PENDING"

	existingPayment := &model.Payment{
		ID:      1,
		OrderID: orderID,
		Status:  string(constants.PaymentStatusPending),
	}

	// Expectations
	mockRepo.On("GetPaymentByOrderID", orderID).Return(existingPayment, nil)
	// UpdatePaymentStatus should NOT be called

	// Execute
	err := svc.RefundPayment(orderID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "cannot refund payment: payment not completed", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestRefundPayment_NotFound_Fails(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	svc := NewPaymentService(mockRepo)

	orderID := "ORD-NOTFOUND"

	// Expectations
	mockRepo.On("GetPaymentByOrderID", orderID).Return(nil, errors.New("record not found"))

	// Execute
	err := svc.RefundPayment(orderID)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
	mockRepo.AssertExpectations(t)
}
