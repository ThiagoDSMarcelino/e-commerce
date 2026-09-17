package main

import (
	"encoding/json"
	"fmt"
)

type ProductRequest struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type Order struct {
	Id       string           `json:"id"`
	Products []ProductRequest `json:"products"`
}

func ParseOrder(data []byte) (*Order, error) {
	var order Order

	err := json.Unmarshal(data, &order)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal message: %v", err)
	}

	return &order, nil
}

func (o *Order) Serialize() ([]byte, error) {
	payload, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal order: %v", err)
	}

	return payload, nil
}
