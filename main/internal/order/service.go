package orders

import (
	"context"
	"errors"
	"fmt"
	repo "main/internal/adapters/postgresql/sqlc"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderErrors string

var (
	ErrOrderWithInvalidProductId = errors.New("invalid product id")
)

type Service interface {
	PlaceOrder(ctx context.Context, data createOrderRequest) (OrderWithItems, error)
	ListOrders(ctx context.Context, page, limit int32) ([]OrderWithItems, error)
	Count(ctx context.Context) (int64, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{repo: repo, db: db}
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
