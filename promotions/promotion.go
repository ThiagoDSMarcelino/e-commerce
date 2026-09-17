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

func (p *Promotion) Serialize() ([]byte, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal promotion: %v", err)
	}

	return payload, nil
}
