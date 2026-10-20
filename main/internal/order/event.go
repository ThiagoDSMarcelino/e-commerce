package orders

import (
	"encoding/json"
	"fmt"
	repo "main/internal/adapters/postgresql/sqlc"
	"uuid"
)

type ProductRequest struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type OrderEvent struct {
	Id          uuid.UUID        `json:"id"`
	Products    []ProductRequest `json:"products"`
	CheckoutUrl string           `json:"checkoutUrl,omitempty"`
}

func newOrderEvent(id uuid.UUID, items []repo.OrderItem) OrderEvent {
	event := OrderEvent{
		Id:       id,
		Products: make([]ProductRequest, 0, len(items)),
	}
	for _, it := range items {
		event.Products = append(event.Products, ProductRequest{
			Id:     it.ProductID,
			Amount: int(it.Quantity),
		})
	}
	return event
}

func parseOrderEvent(data []byte) (*OrderEvent, error) {
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
