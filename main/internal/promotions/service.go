package promotions

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	repo "main/internal/adapters/postgresql/sqlc"
	"main/internal/adapters/rabbitmq"
)

const (
	routingKeyPromotionsRegister   = "interesse.promocao.register"
	routingKeyPromotionsUnregister = "interesse.promocao.unregister"
)

type Service interface {
	ListRegistredEmails(ctx context.Context, data listRegistredEmailQuery) ([]registrationResponse, error)
	Count(ctx context.Context, clientId int64) (int64, error)
	RegisterSubscription(ctx context.Context, data registerSubscriptionRequest) error
	CancelSubscription(ctx context.Context, data cancelSubscriptionRequest) error
}

type svc struct {
	repo *repo.Queries
	rqm  *rabbitmq.Broker
}

func NewService(repo *repo.Queries, rqm *rabbitmq.Broker) Service {
	return &svc{repo: repo, rqm: rqm}
}

func (s *svc) ListRegistredEmails(ctx context.Context, data listRegistredEmailQuery) ([]registrationResponse, error) {
	registrations, err := s.repo.ListRegistredEmail(ctx, repo.ListRegistredEmailParams{
		ClientID: data.ClientID,
		Offset:   (data.Page - 1) * data.Size,
		Limit:    data.Page,
	})
	if err != nil {
		return nil, err
	}

	res := make([]registrationResponse, 0, len(registrations))
	for _, r := range registrations {
		res = append(res, registrationResponse{
			Email:     r.Email,
			CreatedAt: r.CreatedAt.Time,
		})
	}

	return res, nil
}

func (s *svc) Count(ctx context.Context, clientId int64) (int64, error) {
	return s.repo.CountRegistrations(ctx, clientId)
}

func (s *svc) RegisterSubscription(ctx context.Context, data registerSubscriptionRequest) error {
	event := registerPromotionEvent{
		Email:      data.Email,
		ClientID:   data.ClientID,
		Categories: data.Categories,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := s.rqm.Publish(ctx, routingKeyPromotionsRegister, payload); err != nil {
		if !errors.Is(err, rabbitmq.ErrUnroutable) {
			slog.Error("Failed to publish promotion event", "error", err)
			return err
		} else {
			slog.Warn("Order event was not routed to any queue")
		}
	}

	return nil
}

func (s *svc) CancelSubscription(ctx context.Context, data cancelSubscriptionRequest) error {
	event := unregisterPromotionEvent{
		Email:    data.Email,
		ClientID: data.ClientID,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := s.rqm.Publish(ctx, routingKeyPromotionsUnregister, payload); err != nil {
		if !errors.Is(err, rabbitmq.ErrUnroutable) {
			slog.Error("Failed to publish promotion event", "error", err)
			return err
		} else {
			slog.Warn("Order event was not routed to any queue")
		}
	}

	return nil
}
