package products

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrStorageUnavailable = errors.New("storage unavailable")

type Service interface {
	ListAvailable(ctx context.Context) ([]Product, error)
}

type svc struct {
	storageURL string
	client     *http.Client
}

func NewService(storageURL string) Service {
	return &svc{
		storageURL: storageURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *svc) ListAvailable(ctx context.Context) ([]Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.storageURL+"/products", nil)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorageUnavailable, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrStorageUnavailable, res.StatusCode)
	}

	products := []Product{}
	if err := json.NewDecoder(res.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}
