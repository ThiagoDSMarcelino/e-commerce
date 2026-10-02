package orders

import (
	repo "main/internal/adapters/postgresql/sqlc"
	"uuid"
)

func toOrderItem(it repo.OrderItem) OrderItem {
	return OrderItem{
		ProductID: uuid.UUID(it.ProductID.Bytes),
		Quantity:  it.Quantity,
	}
}

func toOrderWithItems(o repo.Order, items []repo.OrderItem) OrderWithItems {
	mapped := make([]OrderItem, len(items))
	for i, it := range items {
		mapped[i] = toOrderItem(it)
	}
	return OrderWithItems{
		ID:        uuid.UUID(o.ID.Bytes),
		ClientID:  o.ClientID,
		CreatedAt: o.CreatedAt.Time,
		Status:    o.Status,
		Items:     mapped,
	}
}
