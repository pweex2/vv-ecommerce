package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

type InventoryClient struct {
	baseURL string
	client  *http.Client
	cb      *gobreaker.CircuitBreaker
}

func NewInventoryClient(url string) *InventoryClient {
	return &InventoryClient{
		baseURL: url,
		client:  NewHTTPClient(2 * time.Second),
		cb:      NewCircuitBreaker("inventory-service"),
	}
}

func (c *InventoryClient) HealthCheck() error {
	resp, err := c.client.Get(c.baseURL + "/health")
	if err != nil {
		return WrapClientError(err, "failed to connect to inventory service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return HandleHTTPError(resp)
	}
	return nil
}

func (c *InventoryClient) Increase(ctx context.Context, sku string, qty int64) error {
	body, _ := json.Marshal(map[string]interface{}{
		"sku":      sku,
		"quantity": qty,
	})

	_, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/inventory/increase", bytes.NewBuffer(body))
		if err != nil {
			return nil, WrapClientError(err, "failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, WrapClientError(err, "failed to connect to inventory service")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, HandleHTTPError(resp)
		}

		return nil, nil
	})

	return err
}

func (c *InventoryClient) Rollback(ctx context.Context, sku string, qty int64, traceID string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"sku":      sku,
		"quantity": qty,
		"trace_id": traceID,
	})

	_, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/inventory/rollback", bytes.NewBuffer(body))
		if err != nil {
			return nil, WrapClientError(err, "failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, WrapClientError(err, "failed to connect to inventory service")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, HandleHTTPError(resp)
		}

		return nil, nil
	})

	return err
}

func (c *InventoryClient) Decrease(ctx context.Context, sku, reqID, orderID, traceID string, qty int64) error {
	body, _ := json.Marshal(map[string]interface{}{
		"sku":        sku,
		"quantity":   qty,
		"request_id": reqID,
		"order_id":   orderID,
		"trace_id":   traceID,
	})

	_, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/inventory/decrease", bytes.NewBuffer(body))
		if err != nil {
			return nil, WrapClientError(err, "failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, WrapClientError(err, "failed to connect to inventory service")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, HandleHTTPError(resp)
		}

		return nil, nil
	})

	return err
}
