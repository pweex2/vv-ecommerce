package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vv-ecommerce/pkg/common/apperror"
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
