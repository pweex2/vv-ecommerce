package repository

import (
	"payment-service/internal/model"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePayment(payment *model.Payment) error
	GetPaymentByOrderID(orderID string) (*model.Payment, error)
	UpdatePaymentStatus(paymentID uint, status string, transactionID string) error
}

type GORMPaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &GORMPaymentRepository{db: db}
}

func (r *GORMPaymentRepository) CreatePayment(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *GORMPaymentRepository) GetPaymentByOrderID(orderID string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *GORMPaymentRepository) UpdatePaymentStatus(paymentID uint, status string, transactionID string) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":         status,
		"transaction_id": transactionID,
	}).Error
}
