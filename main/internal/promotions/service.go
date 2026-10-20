package promotions

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	repo "main/internal/adapters/postgresql/sqlc"
	"main/internal/adapters/rabbitmq"
	"strings"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	routingKeyPromotionsRegister   = "interesse.promocao.register"
	routingKeyPromotionsUnregister = "interesse.promocao.unregister"
)

var ErrRegistrationNotFound = errors.New("inscrição não encontrada")

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
		Limit:    data.Size,
	})
	if err != nil {
		return nil, err
	}

	res := make([]registrationResponse, 0, len(registrations))
	for _, r := range registrations {
		res = append(res, registrationResponse{
			Email:      r.Email,
			Categories: r.Categories,
			CreatedAt:  r.CreatedAt.Time,
		})
	}

	return res, nil
}

func (s *svc) Count(ctx context.Context, clientId int64) (int64, error) {
	return s.repo.CountRegistrations(ctx, clientId)
}

func (s *svc) RegisterSubscription(ctx context.Context, data registerSubscriptionRequest) error {
	email := normalizeEmail(data.Email)

	_, err := s.repo.UpsertRegistration(ctx, repo.UpsertRegistrationParams{
		ID:         pgtype.UUID{Bytes: uuid.NewV7(), Valid: true},
		ClientID:   data.ClientID,
		Email:      email,
		Categories: data.Categories,
	})
	if err != nil {
		return err
	}

	event := registerPromotionEvent{
		Email:      email,
		ClientID:   data.ClientID,
		Categories: data.Categories,
	}

	s.publish(ctx, routingKeyPromotionsRegister, event)

	return nil
}

func (s *svc) CancelSubscription(ctx context.Context, data cancelSubscriptionRequest) error {
	email := normalizeEmail(data.Email)

	rows, err := s.repo.DeleteRegistration(ctx, repo.DeleteRegistrationParams{
		ClientID: data.ClientID,
		Email:    email,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRegistrationNotFound
	}

	event := unregisterPromotionEvent{
		Email:    email,
		ClientID: data.ClientID,
	}

	s.publish(ctx, routingKeyPromotionsUnregister, event)

	return nil
}

// publish envia o evento ao broker. A inscrição já está persistida, então
// falhas aqui são apenas logadas.
func (s *svc) publish(ctx context.Context, routingKey string, event any) {
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to serialize promotion event", "routingKey", routingKey, "error", err)
		return
	}

	if err := s.rqm.Publish(ctx, routingKey, payload); err != nil {
		if !errors.Is(err, rabbitmq.ErrUnroutable) {
			slog.Error("Failed to publish promotion event", "routingKey", routingKey, "error", err)
		} else {
			slog.Warn("Promotion event was not routed to any queue", "routingKey", routingKey)
		}
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
