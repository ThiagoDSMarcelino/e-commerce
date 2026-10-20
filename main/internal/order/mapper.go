package orders

import (
	repo "main/internal/adapters/postgresql/sqlc"
	"time"
	"uuid"
)

func toStatusEvent(o repo.Order) StatusEvent {
	return StatusEvent{
		OrderID:    uuid.UUID(o.ID.Bytes),
		Status:     o.Status,
		UpdatedAt:  time.Now().UTC(),
		PaymentURL: o.PaymentUrl.String,
	}
}

func toOrderItem(it repo.OrderItem) OrderItem {
	return OrderItem{
		ProductID: it.ProductID,
		Quantity:  it.Quantity,
	}
}

func toOrderWithItems(o repo.Order, items []repo.OrderItem) OrderWithItems {
	mapped := make([]OrderItem, len(items))
	for i, it := range items {
		mapped[i] = toOrderItem(it)
	}
	return OrderWithItems{
		ID:         uuid.UUID(o.ID.Bytes),
		ClientID:   o.ClientID,
		CreatedAt:  o.CreatedAt.Time,
		Status:     o.Status,
		Items:      mapped,
		PaymentURL: o.PaymentUrl.String,
	}
}
