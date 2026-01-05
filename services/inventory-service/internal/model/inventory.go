package model

import "time"

type Inventory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `json:"product_id"`
	SKU       string    `gorm:"type:varchar(255);uniqueIndex" json:"sku"` // 添加 SKU 字段并设置为唯一索引
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InventoryDeductionLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                            // 数据库自增 ID
	OrderID   string    `gorm:"type:varchar(255);index" json:"order_id"`         // 业务订单号
	RequestID string    `gorm:"type:varchar(255);uniqueIndex" json:"request_id"` // 幂等性 Key
	SKU       string    `gorm:"type:varchar(255);index" json:"sku"`              // 商品 SKU
	TraceID   string    `gorm:"type:varchar(255);index" json:"trace_id"`         // 分布式追踪 ID
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}
