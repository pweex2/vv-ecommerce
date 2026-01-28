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

// StartWorker starts listening for async rollback tasks
func (c *InventoryCompensator) StartWorker() error {
	return c.mq.Subscribe(c.topic, func(payload []byte) error {
		var msg RollbackMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			return err // Unrecoverable format error, maybe should not retry?
		}

		fmt.Printf("Processing async rollback for SKU %s, Qty %d, TraceID %s\n", msg.SKU, msg.Quantity, msg.TraceID)
		return c.client.Rollback(msg.SKU, msg.Quantity, msg.TraceID)
	})
}
