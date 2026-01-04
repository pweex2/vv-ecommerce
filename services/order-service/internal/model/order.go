package model

type Order struct {
	OrderID     string      `json:"order_id"`
	UserID      int64       `json:"user_id"`
	Status      OrderStatus `json:"status"`
	TotalAmount int64       `json:"total_amount"`
	TraceID     string      `json:"trace_id"`
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
