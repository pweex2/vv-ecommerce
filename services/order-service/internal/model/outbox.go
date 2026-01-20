package model

import (
	"time"

	"gorm.io/datatypes"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

type OutboxEvent struct {
	ID            uint           `gorm:"primaryKey"`
	AggregateType string         `gorm:"size:255;not null"` // e.g., "Order"
	AggregateID   string         `gorm:"size:255;not null"` // e.g., OrderID
	EventType     string         `gorm:"size:255;not null"` // e.g., "InventoryRollback"
	Payload       datatypes.JSON `gorm:"type:json;not null"`
	Status        OutboxStatus   `gorm:"size:50;default:'PENDING';index"`
	TraceID       string         `gorm:"size:255"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
