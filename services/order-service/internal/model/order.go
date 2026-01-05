package model

type Order struct {
	ID          uint        `gorm:"primaryKey" json:"id"`                     // 数据库内部 ID (Primary Key)
	OrderID     string      `gorm:"type:varchar(255);unique" json:"order_id"` // 业务订单 ID (Business Key)
	UserID      int64       `json:"user_id"`
	Status      OrderStatus `json:"status"`
	TotalAmount int64       `json:"total_amount"`
	TraceID     string      `gorm:"type:varchar(255);index" json:"trace_id"` // 追踪 ID
}

type OrderStatus string

const (
	OrderStatusCreated           OrderStatus = "created"
	OrderStatusInventoryReserved OrderStatus = "inventory_reserved"
	OrderStatusPaid              OrderStatus = "paid"
	OrderStatusCompleted         OrderStatus = "completed"
	OrderStatusCancelled         OrderStatus = "cancelled"
	OrderStatusFailed            OrderStatus = "failed"
)
