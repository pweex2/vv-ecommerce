package model

type Order struct {
	OrderID     string `json:"order_id"`
	UserID      int64  `json:"user_id"`
	Status      string `json:"status"`
	TotalAmount int64  `json:"total_amount"`
	TraceID     string `json:"trace_id"`
}
