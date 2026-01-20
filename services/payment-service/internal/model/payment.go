package model

import "time"

type Payment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OrderID       string    `gorm:"type:varchar(255);uniqueIndex" json:"order_id"` // 关联的订单ID
	Amount        int64     `json:"amount"`                                        // 支付金额 (单位：分，避免浮点数)
	Status        string    `json:"status"`                                        // PENDING, COMPLETED, FAILED
	TransactionID string    `json:"transaction_id"`                                // 模拟的交易流水号
	TraceID       string    `gorm:"type:varchar(255);index" json:"trace_id"`       // 全链路追踪ID
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
