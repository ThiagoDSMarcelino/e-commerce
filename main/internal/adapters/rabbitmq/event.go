package rabbitmq

import (
	"encoding/json"
	"fmt"
)

type ProductRequest struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type OrderEvent struct {
	Id       string           `json:"id"`
	Products []ProductRequest `json:"products"`
}

func ParseOrder(data []byte) (*OrderEvent, error) {
	var order OrderEvent

	err := json.Unmarshal(data, &order)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal message: %v", err)
	}

	return &order, nil
}

func (o *OrderEvent) Serialize() ([]byte, error) {
	payload, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal order: %v", err)
	}

	return payload, nil
}
