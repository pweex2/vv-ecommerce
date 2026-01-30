package handler

import (
	"context"
	"encoding/json"
	"log"

	"inventory-service/internal/service"
	"vv-ecommerce/pkg/async"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type EventHandler struct {
	service *service.InventoryService
	mq      async.MessageQueue
}

func NewEventHandler(service *service.InventoryService, mq async.MessageQueue) *EventHandler {
	return &EventHandler{
		service: service,
		mq:      mq,
	}
}

func (h *EventHandler) RegisterSubscribers() error {
	// 在这里注册所有的订阅
	return h.mq.Subscribe("inventory_rollback", h.HandleInventoryRollback)
}

func (h *EventHandler) HandleInventoryRollback(payload []byte, traceHeaders map[string]string) error {
	// Extract trace context
	ctx := context.Background()
	carrier := propagation.MapCarrier(traceHeaders)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Start a new span
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "async_process_inventory_rollback",
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	var msg struct {
		SKU      string `json:"sku"`
		Quantity int64  `json:"quantity"`
		TraceID  string `json:"trace_id"`
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		span.RecordError(err)
		log.Printf("Failed to unmarshal rollback message: %v", err)
		return nil // Don't retry malformed messages
	}

	log.Printf("Processing rollback for SKU: %s, Qty: %d, TraceID: %s", msg.SKU, msg.Quantity, msg.TraceID)

	// Call service
	// Note: RollbackInventory expects 'int' for quantity, msg has int64
	err := h.service.RollbackInventory(ctx, msg.SKU, int(msg.Quantity), msg.TraceID)
	if err != nil {
		span.RecordError(err)
		log.Printf("Failed to rollback inventory: %v", err)
		return err // Retry
	}
	return nil
}
