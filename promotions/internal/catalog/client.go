package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

var ErrStorageUnavailable = errors.New("storage unavailable")

type Product struct {
	ID       string
	Name     string
	Amount   int
	Category Category
}

type storageProduct struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type Client struct {
	storageURL string
	client     *http.Client

	mu         sync.Mutex
	categories map[string]Category
}

func NewClient(storageURL string) *Client {
	return &Client{
		storageURL: storageURL,
		client:     &http.Client{Timeout: 5 * time.Second},
		categories: make(map[string]Category),
	}
}

func (c *Client) Products(ctx context.Context) ([]Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.storageURL+"/products", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorageUnavailable, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrStorageUnavailable, res.StatusCode)
	}

	var items []storageProduct
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	products := make([]Product, 0, len(items))
	for _, item := range items {
		products = append(products, Product{
			ID:       item.ID,
			Name:     item.Name,
			Amount:   item.Amount,
			Category: c.categoryOf(item.ID),
		})
	}

	return products, nil
}

func (c *Client) categoryOf(productID string) Category {
	c.mu.Lock()
	defer c.mu.Unlock()

	category, ok := c.categories[productID]
	if !ok {
		category = Categories[rand.IntN(len(Categories))]
		c.categories[productID] = category
	}

	return category
}
