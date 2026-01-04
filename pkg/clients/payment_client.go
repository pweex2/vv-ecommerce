package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type PaymentClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentClient(url string) *PaymentClient {
	return &PaymentClient{
		baseURL: url,
		client:  &http.Client{Timeout: 5 * time.Second},
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

func (c *PaymentClient) ProcessPayment(orderID string, amount int64) (*PaymentResponse, error) {
	reqBody := PaymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(
		c.baseURL+"/payments",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 尝试读取错误信息
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return nil, errors.New(errResp.Error)
		}
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &paymentResp, nil
}

func (c *PaymentClient) GetPayment(orderID string) (*PaymentResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/payments?order_id=%s", c.baseURL, orderID))
	if err != nil {
		return nil, fmt.Errorf("failed to call payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &paymentResp, nil
}
