package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ChargeClient struct {
	mockURL    string
	webhookURL string
	http       *http.Client
}

type chargeRequest struct {
	OrderId    string `json:"orderId"`
	WebhookUrl string `json:"webhookUrl"`
}

type chargeResponse struct {
	CheckoutUrl string `json:"checkoutUrl"`
}

func NewChargeClient(settings *Settings) *ChargeClient {
	return &ChargeClient{
		mockURL:    settings.PaymentMockURL,
		webhookURL: settings.WebhookURL,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *ChargeClient) CreateCharge(ctx context.Context, orderId string) (string, error) {
	body, err := json.Marshal(chargeRequest{OrderId: orderId, WebhookUrl: c.webhookURL})
	if err != nil {
		return "", fmt.Errorf("Failed to marshal charge: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.mockURL+"/charges", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("Failed to create charge request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to request charge: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("Payment mock answered with status %d", res.StatusCode)
	}

	var charge chargeResponse
	if err := json.NewDecoder(res.Body).Decode(&charge); err != nil {
		return "", fmt.Errorf("Failed to decode charge response: %v", err)
	}

	if charge.CheckoutUrl == "" {
		return "", errors.New("Payment mock answered without a checkout url")
	}

	return charge.CheckoutUrl, nil
}
