package service

import (
	"encoding/json"
	"fmt"
	"vv-ecommerce/pkg/async"
	"vv-ecommerce/pkg/clients"
)

type RollbackMessage struct {
	SKU      string `json:"sku"`
	Quantity int64  `json:"quantity"`
	TraceID  string `json:"trace_id"`
}

type InventoryCompensator struct {
	client *clients.InventoryClient
	mq     async.MessageQueue
	topic  string
}

func NewInventoryCompensator(client *clients.InventoryClient, mq async.MessageQueue) *InventoryCompensator {
	return &InventoryCompensator{
		client: client,
		mq:     mq,
		topic:  "inventory_rollback",
	}
}

// Compensate tries to rollback synchronously. If it fails, it pushes the task to MQ.
func (c *InventoryCompensator) Compensate(sku string, quantity int64, traceID string) {
	// 1. Try synchronous rollback
	err := c.client.Increase(sku, quantity, traceID)
	if err == nil {
		return
	}

	fmt.Printf("Sync rollback failed for SKU %s: %v. Queueing async compensation...\n", sku, err)

	// 2. If failed, push to MQ
	msg := RollbackMessage{
		SKU:      sku,
		Quantity: quantity,
		TraceID:  traceID,
	}
	payload, _ := json.Marshal(msg) // Ignore marshal error for struct
	if err := c.mq.Publish(c.topic, payload); err != nil {
		// Critical: If MQ also fails, we are in trouble (Data inconsistency risk)
		// In production, log to a persistent local file or alert system
		fmt.Printf("CRITICAL: Failed to publish rollback message: %v\n", err)
	}
}

// StartWorker starts listening for async rollback tasks
func (c *InventoryCompensator) StartWorker() error {
	return c.mq.Subscribe(c.topic, func(payload []byte) error {
		var msg RollbackMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			return err // Unrecoverable format error, maybe should not retry?
		}

		fmt.Printf("Processing async rollback for SKU %s, Qty %d, TraceID %s\n", msg.SKU, msg.Quantity, msg.TraceID)
		return c.client.Increase(msg.SKU, msg.Quantity, msg.TraceID)
	})
}