package service

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/model"
	"order-service/internal/repository"
	"time"

	"vv-ecommerce/pkg/async"
)

type OutboxProcessor struct {
	repo        repository.OrderRepository
	queue       async.MessageQueue
	interval    time.Duration
	stopChan    chan struct{}
}

func NewOutboxProcessor(repo repository.OrderRepository, queue async.MessageQueue) *OutboxProcessor {
	return &OutboxProcessor{
		repo:     repo,
		queue:    queue,
		interval: 5 * time.Second, // Poll every 5 seconds
		stopChan: make(chan struct{}),
	}
}

func (p *OutboxProcessor) Start() {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.processEvents()
			case <-p.stopChan:
				log.Println("OutboxProcessor stopping...")
				return
			}
		}
	}()
}

func (p *OutboxProcessor) Stop() {
	close(p.stopChan)
}

func (p *OutboxProcessor) processEvents() {
	ctx := context.Background() // Should ideally have a timeout

	// 1. Fetch pending events
	events, err := p.repo.GetPendingOutboxEvents(ctx, 10) // Batch size 10
	if err != nil {
		log.Printf("Error fetching outbox events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		// 2. Process based on EventType
		switch event.EventType {
		case "InventoryRollback":
			if err := p.publishInventoryRollback(ctx, event); err != nil {
				log.Printf("Error processing event %d: %v", event.ID, err)
				// Retry strategy? For now, leave as PENDING to be picked up again.
				// In production, might want backoff or FAILED status after N attempts.
			} else {
				// 3. Mark as PROCESSED
				if err := p.repo.UpdateOutboxEventStatus(ctx, event.ID, model.OutboxStatusProcessed); err != nil {
					log.Printf("Error updating event status %d: %v", event.ID, err)
				}
			}
		default:
			log.Printf("Unknown event type: %s", event.EventType)
			p.repo.UpdateOutboxEventStatus(ctx, event.ID, model.OutboxStatusFailed)
		}
	}
}

func (p *OutboxProcessor) publishInventoryRollback(ctx context.Context, event model.OutboxEvent) error {
	var payload struct {
		SKU      string `json:"sku"`
		Quantity int64  `json:"quantity"`
		TraceID  string `json:"trace_id"`
	}

	// datatypes.JSON is []byte alias
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	// Construct message for MQ
	message := map[string]interface{}{
		"sku":      payload.SKU,
		"quantity": payload.Quantity,
		"trace_id": payload.TraceID,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Publish to RabbitMQ
	return p.queue.Publish("inventory_rollback", messageBytes)
}
