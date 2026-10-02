package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	repo "main/internal/adapters/postgresql/sqlc"
	"main/internal/adapters/rabbitmq"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	routingKeyCreated = "pedido.criado"
	routingKeyDeleted = "pedido.excluido"
)

type OrderErrors string

var (
	ErrOrderWithInvalidProductId = errors.New("invalid product id")
)

type Service interface {
	PlaceOrder(ctx context.Context, data createOrderRequest) (OrderWithItems, error)
	ListOrders(ctx context.Context, page, limit int32) ([]OrderWithItems, error)
	Count(ctx context.Context) (int64, error)
	HandleStatusEvent(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
	rqm  *rabbitmq.Broker
}

func NewService(repo *repo.Queries, db *pgx.Conn, rqm *rabbitmq.Broker) Service {
	return &svc{repo: repo, db: db, rqm: rqm}
}

func (s *svc) PlaceOrder(ctx context.Context, data createOrderRequest) (OrderWithItems, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrderWithItems{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	id := pgtype.UUID{Bytes: uuid.NewV7(), Valid: true}
	createdOrder, err := qtx.CreateOrder(ctx, repo.CreateOrderParams{
		ID:       id,
		ClientID: data.ClientID,
		Status:   string(StatusCreated),
	})
	if err != nil {
		return OrderWithItems{}, err
	}

	items := make([]repo.OrderItem, 0, len(data.Items))
	for _, item := range data.Items {
		var productID pgtype.UUID
		if err = productID.Scan(item.ProductID); err != nil {
			return OrderWithItems{}, fmt.Errorf("invalid product id %s: %w", item.ProductID, ErrOrderWithInvalidProductId)
		}

		createdItem, err := s.repo.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:   id,
			ProductID: productID,
			Quantity:  item.Quantity,
		})
		if err != nil {
			return OrderWithItems{}, err
		}
		items = append(items, createdItem)
	}

	if err := tx.Commit(ctx); err != nil {
		return OrderWithItems{}, err
	}

	event := OrderEvent{
		Id:       createdOrder.ID.Bytes,
		Products: make([]ProductRequest, 0, len(items)),
	}
	for _, it := range items {
		event.Products = append(event.Products, ProductRequest{
			Id:     uuid.UUID(it.ProductID.Bytes).String(),
			Amount: int(it.Quantity),
		})
	}
	payload, err := event.Serialize()

	if err := s.rqm.Publish(ctx, routingKeyCreated, payload); err != nil {
		if !errors.Is(err, rabbitmq.ErrUnroutable) {
			slog.Error("Failed to publish order", "id", createdOrder.ID, "error", err)
		} else {
			slog.Warn("Order event was not routed to any queue", "id", createdOrder.ID)
		}
	}

	return toOrderWithItems(createdOrder, items), nil
}

func (s *svc) ListOrders(ctx context.Context, page, limit int32) ([]OrderWithItems, error) {
	orders, err := s.repo.ListOrders(ctx, repo.ListOrdersParams{
		Offset: (page - 1) * limit,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return []OrderWithItems{}, nil
	}

	ids := make([]pgtype.UUID, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
	}

	items, err := s.repo.ListOrderItemsByOrderIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	itemsByOrder := make(map[[16]byte][]repo.OrderItem, len(orders))
	for _, it := range items {
		itemsByOrder[it.OrderID.Bytes] = append(itemsByOrder[it.OrderID.Bytes], it)
	}

	result := make([]OrderWithItems, len(orders))
	for i, o := range orders {
		result[i] = toOrderWithItems(o, itemsByOrder[o.ID.Bytes])
	}

	return result, nil
}

func (s *svc) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *svc) HandleStatusEvent(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse {
	order, err := parseOrderEvent(event.Data)
	if err != nil {
		slog.Error("Failed to parse order", "routingKey", event.RoutingKey, "error", err)
		return rabbitmq.Rejected
	}

	status, ok := statusByRoutingKey[event.RoutingKey]
	if !ok {
		slog.Warn("Received event with unknown routing key",
			"routingKey", event.RoutingKey, "order_id", order.Id)
		return rabbitmq.Rejected
	}

	if err := s.updateOrderStatus(ctx, order.Id, status); err != nil {
		return rabbitmq.Requeued
	}

	slog.Info("Order status updated", "order_id", order.Id, "status", status)

	if !cancelsOrder[event.RoutingKey] {
		return rabbitmq.Accepted
	}

	payload, err := json.Marshal(order)
	if err != nil {
		return rabbitmq.Requeued
	}

	if err := s.rqm.Publish(ctx, routingKeyDeleted, payload); err != nil {
		if !errors.Is(err, rabbitmq.ErrUnroutable) {
			slog.Error("Failed to publish cancellation", "order_id", order.Id, "error", err)
			return rabbitmq.Requeued
		}

		slog.Warn("Cancellation event was not routed to any queue", "id", order.Id)
	}

	slog.Info("order cancelled", "id", order.Id)

	return rabbitmq.Accepted
}

func (s *svc) updateOrderStatus(ctx context.Context, orderId uuid.UUID, status OrderStatus) error {
	return s.repo.UpdateOrderStatus(ctx, repo.UpdateOrderStatusParams{
		OrderID: pgtype.UUID{Bytes: orderId, Valid: true},
		Status:  string(status),
	})
}
