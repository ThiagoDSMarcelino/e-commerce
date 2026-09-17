package main

import (
	"encoding/json"
	"fmt"
)

type Promotion struct {
	Category   string `json:"category"`
	Discount   int    `json:"discount"`
	ValidUntil string `json:"validUntil"`
}

func ParsePromotion(data []byte) (*Promotion, error) {
	var promotion Promotion

	err := json.Unmarshal(data, &promotion)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal message: %v", err)
	}

	return &promotion, nil
}
