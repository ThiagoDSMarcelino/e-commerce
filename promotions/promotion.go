package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Product struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Catalog struct {
	Products []Product `json:"products"`
}

type Promotion struct {
	ProductId   string `json:"productId"`
	ProductName string `json:"productName"`
	Category    string `json:"category"`
	Discount    int    `json:"discount"`
	ValidUntil  string `json:"validUntil"`
}

func LoadCatalog(path string) ([]Product, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s: %v", path, err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("Failed to parse %s: %v", path, err)
	}

	if len(catalog.Products) == 0 {
		return nil, fmt.Errorf("%s has no products", path)
	}

	return catalog.Products, nil
}

func (p *Promotion) Serialize() ([]byte, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal promotion: %v", err)
	}

	return payload, nil
}
