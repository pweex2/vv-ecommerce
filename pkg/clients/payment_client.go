package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vv-ecommerce/pkg/common/apperror"

	"github.com/sony/gobreaker"
)

type PaymentClient struct {
	baseURL string
	client  *http.Client
	cb      *gobreaker.CircuitBreaker
}

func NewPaymentClient(url string) *PaymentClient {
	return &PaymentClient{
		baseURL: url,
		client:  NewHTTPClient(5 * time.Second),
		cb:      NewCircuitBreaker("payment-service"),
	}
}

type PaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type PaymentResponse struct {
	ID            uint      `json:"id"`
	OrderID       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	TransactionID string    `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (c *PaymentClient) ProcessPayment(ctx context.Context, orderID string, amount int64, traceID string) (*PaymentResponse, error) {
	reqBody := PaymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}
	body, _ := json.Marshal(reqBody)

	// Execute with circuit breaker
	res, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/payments", bytes.NewBuffer(body))
		if err != nil {
			return nil, WrapClientError(err, "failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")
		if traceID != "" {
			req.Header.Set("X-Trace-ID", traceID)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, WrapClientError(err, "failed to call payment service")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, HandleHTTPError(resp)
		}

		var paymentResp PaymentResponse
		if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
			return nil, apperror.Internal("failed to decode payment response", err)
		}

		return &paymentResp, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*PaymentResponse), nil
}

func (c *PaymentClient) Refund(ctx context.Context, orderID string, traceID string) error {
	reqBody := map[string]string{"order_id": orderID}
	body, _ := json.Marshal(reqBody)

	// Execute with circuit breaker
	_, err := c.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/payments/refund", bytes.NewBuffer(body))
		if err != nil {
			return nil, WrapClientError(err, "failed to create request")
		}
		req.Header.Set("Content-Type", "application/json")
		if traceID != "" {
			req.Header.Set("X-Trace-ID", traceID)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, WrapClientError(err, "failed to call payment service")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, HandleHTTPError(resp)
		}

		return nil, nil
	})

	return err
}

func (c *PaymentClient) GetPayment(orderID string) (*PaymentResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/payments?order_id=%s", c.baseURL, orderID))
	if err != nil {
		return nil, WrapClientError(err, "failed to call payment service")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.NotFound("payment not found", nil)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, HandleHTTPError(resp)
	}

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, apperror.Internal("failed to decode payment response", err)
	}

	return &paymentResp, nil
}
