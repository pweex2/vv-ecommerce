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
	RequestID string    `gorm:"type:varchar(255);uniqueIndex" json:"request_id"` // 添加 RequestID 字段并设置为唯一索引
	SKU       string    `gorm:"type:varchar(255);index" json:"sku"`              // 添加 SKU 字段并设置索引
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}
