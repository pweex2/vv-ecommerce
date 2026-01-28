package service

import (
	"errors"
	"payment-service/internal/model"
	"payment-service/internal/repository"
	"time"
	"vv-ecommerce/pkg/common/constants"

	"github.com/google/uuid"
)

type PaymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) ProcessPayment(orderID string, amount int64) (*model.Payment, error) {
	// 1. 创建初始支付记录 (PENDING)
	payment := &model.Payment{
		OrderID: orderID,
		Amount:  amount,
		Status:  string(constants.PaymentStatusPending),
	}
	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, err
	}

	// 2. 模拟调用第三方支付网关 (如 Stripe/PayPal)
	// 这里我们模拟一个处理延迟和随机失败
	// 实际生产中，这里会是调用外部 API
	time.Sleep(500 * time.Millisecond) // 模拟网络延迟

	// 简单的模拟逻辑：金额为负数直接失败，否则大概率成功
	// 在真实场景中，可能会有更复杂的风控检查
	var newStatus string
	var transactionID string
	var err error

	if amount < 0 {
		newStatus = string(constants.PaymentStatusFailed)
		err = errors.New("invalid amount")
	} else if amount == 9999 {
		// 模拟特定金额触发支付失败 (用于测试分布式事务回滚)
		newStatus = string(constants.PaymentStatusFailed)
		err = errors.New("simulated payment failure for testing")
	} else {
		// 模拟 10% 的失败率 (可选，用于测试容错)
		// if rand.Intn(10) == 0 {
		// 	newStatus = string(constants.PaymentStatusFailed)
		// 	err = errors.New("payment gateway rejected")
		// } else {
		newStatus = string(constants.PaymentStatusCompleted)
		transactionID = uuid.New().String()
		// }
	}

	// 3. 更新支付状态
	if updateErr := s.repo.UpdatePaymentStatus(payment.ID, newStatus, transactionID); updateErr != nil {
		// 如果更新数据库失败，这是一个严重错误（数据不一致）
		// 实际场景可能需要异步重试或人工介入
		return nil, updateErr
	}

	payment.Status = newStatus
	payment.TransactionID = transactionID

	return payment, err
}

func (s *PaymentService) GetPayment(orderID string) (*model.Payment, error) {
	return s.repo.GetPaymentByOrderID(orderID)
}
