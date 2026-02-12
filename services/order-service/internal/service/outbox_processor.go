package service

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/model"
	"order-service/internal/repository"
	"time"

	"vv-ecommerce/pkg/async"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type OutboxProcessor struct {
	repo     repository.OrderRepository
	queue    async.MessageQueue
	interval time.Duration
	stopChan chan struct{}
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

	// 1. Fetch pending events with lock (SKIP LOCKED)
	// We use a timeout to prevent hanging DB connections
	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	events, err := p.repo.FetchAndLockPendingEvents(fetchCtx, 10) // Batch size 10
	if err != nil {
		log.Printf("Error fetching outbox events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		// 2. Process based on EventType
		var processErr error
		switch event.EventType {
		case "InventoryRollback":
			processErr = p.publishInventoryRollback(ctx, event)
		default:
			log.Printf("Unknown event type: %s", event.EventType)
			// Mark as FAILED immediately for unknown types
			p.repo.UpdateOutboxEventStatus(ctx, event.ID, model.OutboxStatusFailed, "Unknown event type")
			continue
		}

		if processErr != nil {
			log.Printf("Error processing event %d (Attempt %d): %v", event.ID, event.RetryCount+1, processErr)
			
			// Max retries logic (e.g., 3 times)
			const maxRetries = 3
			if event.RetryCount >= maxRetries {
				log.Printf("Event %d reached max retries, marking as FAILED", event.ID)
				p.repo.UpdateOutboxEventStatus(ctx, event.ID, model.OutboxStatusFailed, processErr.Error())
			} else {
				// Increment retry count and keep as PENDING (or ideally, set a 'NextRetryAt' time)
				// For now, simple increment.
				p.repo.IncrementRetryCount(ctx, event.ID, processErr.Error())
			}
		} else {
			// 3. Mark as PROCESSED
			if err := p.repo.UpdateOutboxEventStatus(ctx, event.ID, model.OutboxStatusProcessed, ""); err != nil {
				log.Printf("Error updating event status %d: %v", event.ID, err)
			}
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

	// Extract TraceID from event (preferred) or payload
	traceID := event.TraceID
	if traceID == "" {
		traceID = payload.TraceID
	}

	// Prepare trace headers
	traceHeaders := make(map[string]string)

	// Attempt to reconstruct trace context
	if len(traceID) == 32 { // Valid OTEL TraceID length
		tid, err := trace.TraceIDFromHex(traceID)
		if err == nil {
			// We have the TraceID from the database (persisted from the original request),
			// but we lost the parent SpanID because it wasn't stored.
			// To maintain trace continuity in Jaeger, we construct a RemoteSpanContext
			// using the persisted TraceID and a generated "dummy" SpanID.
			// This makes the consumer's span appear as a child of this dummy span,
			// linking it to the original trace.
			dummySpanID := [8]byte{1, 2, 3, 4, 5, 6, 7, 8} // Deterministic dummy parent ID

			sc := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    tid,
				SpanID:     dummySpanID,
				TraceFlags: trace.FlagsSampled,
				Remote:     true,
			})

			// Create a context with this remote span context
			ctxWithSpan := trace.ContextWithRemoteSpanContext(ctx, sc)

			// Inject into headers for propagation
			otel.GetTextMapPropagator().Inject(ctxWithSpan, propagation.MapCarrier(traceHeaders))
		}
	}

	// Construct message for MQ
	message := map[string]interface{}{
		"sku":      payload.SKU,
		"quantity": payload.Quantity,
		"trace_id": traceID,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Publish to RabbitMQ with headers
	return p.queue.Publish("inventory_rollback", messageBytes, traceHeaders)
}
