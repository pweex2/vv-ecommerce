package clients

import (
	"bytes"
	"encoding/json"
	"errors"
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
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("inventory health check failed")
	}
	return nil
}

func (c *InventoryClient) Decrease(sku, reqID string, qty int64) error {
	body, _ := json.Marshal(map[string]interface{}{
		"sku":        sku,
		"quantity":   qty,
		"request_id": reqID,
	})

	resp, err := c.client.Post(
		c.baseURL+"/inventory/decrease",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		// 尝试读取错误信息
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return errors.New(errResp.Error)
		}
		return errors.New("inventory decrease failed")
	}

	return nil
}
