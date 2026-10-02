package orders

import (
	"context"
	repo "main/internal/adapters/postgresql/sqlc"
)

type Service interface {
	ListProducts(ctx context.Context, page, limit int32) ([]repo.Order, error)
	Count(ctx context.Context) (int64, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) ListProducts(ctx context.Context, page, limit int32) ([]repo.Order, error) {
	list, err := s.repo.ListOrders(ctx, repo.ListOrdersParams{
		Offset: (page - 1) * limit,
		Limit:  limit,
	})

	if list == nil {
		list = []repo.Order{}
	}

	return list, err
}

func (s *svc) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
