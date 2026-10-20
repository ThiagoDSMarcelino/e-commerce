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
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	routingKeyCreated        = "pedido.criado"
	routingKeyDeleted        = "pedido.excluido"
	routingKeyPaymentPending = "pagamento.pendente"
)

var (
	ErrOrderNotFound       = errors.New("pedido não encontrado")
	ErrOrderNotCancellable = errors.New("pedido não pode ser cancelado")
)

type Service interface {
	PlaceOrder(ctx context.Context, data createOrderRequest) (OrderWithItems, error)
	ListOrders(ctx context.Context, data listOrdersQuery) ([]OrderWithItems, error)
	Count(ctx context.Context, clientId int64) (int64, error)
	CancelOrder(ctx context.Context, orderId uuid.UUID, clientId int64) (OrderWithItems, error)
	HandleStatusEvent(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse
	Subscribe(clientId int64) (<-chan StatusEvent, func())
}

type svc struct {
	repo   *repo.Queries
	db     *pgxpool.Pool
	rqm    *rabbitmq.Broker
	events *Broadcaster
}

func NewService(repo *repo.Queries, db *pgxpool.Pool, rqm *rabbitmq.Broker, events *Broadcaster) Service {
	return &svc{repo: repo, db: db, rqm: rqm, events: events}
}

func (s *svc) Subscribe(clientId int64) (<-chan StatusEvent, func()) {
	return s.events.Subscribe(clientId)
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
		createdItem, err := qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:   id,
			ProductID: item.ProductID,
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

	event := newOrderEvent(createdOrder.ID.Bytes, items)
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

func (s *svc) ListOrders(ctx context.Context, data listOrdersQuery) ([]OrderWithItems, error) {
	orders, err := s.repo.ListOrders(ctx, repo.ListOrdersParams{
		ClientID: data.ClientID,
		Offset:   (data.Page - 1) * data.Size,
		Limit:    data.Size,
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

func (s *svc) Count(ctx context.Context, clientId int64) (int64, error) {
	return s.repo.CountOrders(ctx, clientId)
}

func (s *svc) CancelOrder(ctx context.Context, orderId uuid.UUID, clientId int64) (OrderWithItems, error) {
	id := pgtype.UUID{Bytes: orderId, Valid: true}

	cancelled, err := s.repo.CancelOrder(ctx, repo.CancelOrderParams{
		Status:         string(StatusCancelled),
		OrderID:        id,
		ClientID:       clientId,
		NonCancellable: nonCancellableStatuses,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		current, err := s.repo.GetOrder(ctx, repo.GetOrderParams{OrderID: id, ClientID: clientId})
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderWithItems{}, ErrOrderNotFound
		}
		if err != nil {
			return OrderWithItems{}, err
		}
		return OrderWithItems{}, fmt.Errorf("%w: status %q", ErrOrderNotCancellable, current.Status)
	}
	if err != nil {
		return OrderWithItems{}, err
	}

	items, err := s.repo.ListOrderItemsByOrderIDs(ctx, []pgtype.UUID{id})
	if err != nil {
		return OrderWithItems{}, err
	}

	// O storage devolve ao estoque o que tiver sido reservado para o pedido.
	event := newOrderEvent(orderId, items)
	payload, err := event.Serialize()
	if err != nil {
		return OrderWithItems{}, err
	}

	if err := s.rqm.Publish(ctx, routingKeyDeleted, payload); err != nil {
		if !errors.Is(err, rabbitmq.ErrUnroutable) {
			slog.Error("Failed to publish cancellation", "order_id", orderId, "error", err)
		} else {
			slog.Warn("Cancellation event was not routed to any queue", "id", orderId)
		}
	}

	slog.Info("order cancelled by client", "id", orderId, "client_id", clientId)

	return toOrderWithItems(cancelled, items), nil
}

func (s *svc) HandleStatusEvent(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse {
	order, err := parseOrderEvent(event.Data)
	if err != nil {
		slog.Error("Failed to parse order", "routingKey", event.RoutingKey, "error", err)
		return rabbitmq.Rejected
	}

	// pagamento.pendente não muda o status: só traz o link de pagamento.
	if event.RoutingKey == routingKeyPaymentPending {
		return s.savePaymentUrl(ctx, order)
	}

	status, ok := statusByRoutingKey[event.RoutingKey]
	if !ok {
		slog.Warn("Received event with unknown routing key",
			"routingKey", event.RoutingKey, "order_id", order.Id)
		return rabbitmq.Rejected
	}

	updated, err := s.updateOrderStatus(ctx, order.Id, status)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Info("Ignoring status event for cancelled or unknown order",
			"order_id", order.Id, "routingKey", event.RoutingKey)
		return rabbitmq.Accepted
	}
	if err != nil {
		slog.Error("Failed to update order status", "order_id", order.Id, "error", err)
		return rabbitmq.Requeued
	}

	slog.Info("Order status updated", "order_id", order.Id, "status", status)

	s.events.Publish(updated.ClientID, toStatusEvent(updated))

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

// savePaymentUrl guarda o link do mock de pagamento, que a listagem devolve
// como paymentUrl para a UI mostrar o botão "Pagar".
func (s *svc) savePaymentUrl(ctx context.Context, order *OrderEvent) rabbitmq.MessageResponse {
	if order.CheckoutUrl == "" {
		slog.Error("Payment pending event without checkout url", "order_id", order.Id)
		return rabbitmq.Rejected
	}

	updated, err := s.repo.SetOrderPaymentUrl(ctx, repo.SetOrderPaymentUrlParams{
		PaymentUrl:      order.CheckoutUrl,
		OrderID:         pgtype.UUID{Bytes: order.Id, Valid: true},
		CancelledStatus: string(StatusCancelled),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Info("Ignoring payment url for cancelled or unknown order", "order_id", order.Id)
		return rabbitmq.Accepted
	}
	if err != nil {
		slog.Error("Failed to save payment url", "order_id", order.Id, "error", err)
		return rabbitmq.Requeued
	}

	slog.Info("Order payment url saved", "order_id", order.Id, "payment_url", order.CheckoutUrl)

	s.events.Publish(updated.ClientID, toStatusEvent(updated))

	return rabbitmq.Accepted
}

func (s *svc) updateOrderStatus(ctx context.Context, orderId uuid.UUID, status OrderStatus) (repo.Order, error) {
	return s.repo.UpdateOrderStatus(ctx, repo.UpdateOrderStatusParams{
		OrderID:         pgtype.UUID{Bytes: orderId, Valid: true},
		Status:          string(status),
		CancelledStatus: string(StatusCancelled),
	})
}
