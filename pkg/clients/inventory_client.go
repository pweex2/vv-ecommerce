package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	pb "vv-ecommerce/pkg/proto/inventory"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient struct {
	baseURL    string
	grpcTarget string
	client     *http.Client
	grpcConn   *grpc.ClientConn
	grpcClient pb.InventoryServiceClient
	cb         *gobreaker.CircuitBreaker
}

func NewInventoryClient(url string, grpcTarget string) *InventoryClient {
	// Initialize gRPC connection (Lazy connection)
	// In production, you might want to manage this connection lifecycle more carefully
	var conn *grpc.ClientConn
	var grpcClient pb.InventoryServiceClient
	var err error

	if grpcTarget != "" {
		conn, err = grpc.NewClient(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			grpcClient = pb.NewInventoryServiceClient(conn)
		}
	}

	return &InventoryClient{
		baseURL:    url,
		grpcTarget: grpcTarget,
		client:     NewHTTPClient(2 * time.Second),
		grpcConn:   conn,
		grpcClient: grpcClient,
		cb:         NewCircuitBreaker("inventory-service"),
	}
}

func (c *InventoryClient) Close() {
	if c.grpcConn != nil {
		c.grpcConn.Close()
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
	// Try gRPC first if available
	if c.grpcClient != nil {
		_, err := c.cb.Execute(func() (interface{}, error) {
			req := &pb.DecreaseStockRequest{
				Sku:       sku,
				Quantity:  qty,
				RequestId: reqID,
				OrderId:   orderID,
				TraceId:   traceID,
			}
			resp, err := c.grpcClient.DecreaseStock(ctx, req)
			if err != nil {
				return nil, err
			}
			if !resp.Success {
				return nil, WrapClientError(nil, resp.Message)
			}
			return resp, nil
		})
		if err == nil {
			return nil
		}
		// If gRPC fails, we could fallback to HTTP, or just return error.
		// For this migration, let's stick to gRPC error if configured.
		// To be robust, we return the error here.
		return err
	}

	// Fallback to HTTP (Old Logic)
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
