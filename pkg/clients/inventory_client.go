package clients

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type InventoryClient struct {
	baseURL string
	client  *http.Client
}

func NewInventoryClient(url string) *InventoryClient {
	return &InventoryClient{
		baseURL: url,
		client:  &http.Client{Timeout: 2 * time.Second},
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

func (c *InventoryClient) Increase(sku string, qty int64, traceID string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"sku":      sku,
		"quantity": qty,
		"trace_id": traceID,
	})

	resp, err := c.client.Post(
		c.baseURL+"/inventory/increase",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return WrapClientError(err, "failed to connect to inventory service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return HandleHTTPError(resp)
	}

	return nil
}

func (c *InventoryClient) Decrease(sku, reqID, orderID, traceID string, qty int64) error {
	body, _ := json.Marshal(map[string]interface{}{
		"sku":        sku,
		"quantity":   qty,
		"request_id": reqID,
		"order_id":   orderID,
		"trace_id":   traceID,
	})

	resp, err := c.client.Post(
		c.baseURL+"/inventory/decrease",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return WrapClientError(err, "failed to connect to inventory service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return HandleHTTPError(resp)
	}

	return nil
}
