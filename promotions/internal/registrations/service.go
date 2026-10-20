package registrations

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"sync"
)

type Service struct {
	mu          sync.RWMutex
	subscribers map[subscriberKey]Subscriber
}

type subscriberKey struct {
	clientID int64
	email    string
}

type Subscriber struct {
	Email      string
	ClientID   int64
	Categories []string
}

func NewService() *Service {
	return &Service{
		subscribers: make(map[subscriberKey]Subscriber),
	}
}

func (s *Service) RegisterEmail(ctx context.Context, event RegisterPromotionEvent) error {
	email := normalize(event.Email)
	if email == "" {
		slog.Warn("Ignoring registration without email", "clientId", event.ClientID)
		return nil
	}

	categories := make([]string, 0, len(event.Categories))
	for _, category := range event.Categories {
		if c := normalize(category); c != "" && !slices.Contains(categories, c) {
			categories = append(categories, c)
		}
	}

	s.mu.Lock()
	s.subscribers[subscriberKey{event.ClientID, email}] = Subscriber{
		Email:      email,
		ClientID:   event.ClientID,
		Categories: categories,
	}
	s.mu.Unlock()

	slog.Info("Registered for promotions", "email", email, "clientId", event.ClientID, "categories", categories)
	return nil
}

func (s *Service) UnregisterEmail(ctx context.Context, event UnregisterPromotionEvent) error {
	email := normalize(event.Email)

	s.mu.Lock()
	delete(s.subscribers, subscriberKey{event.ClientID, email})
	s.mu.Unlock()

	slog.Info("Unregistered from promotions", "email", email, "clientId", event.ClientID)
	return nil
}

func (s *Service) Recipients(category string) []string {
	category = normalize(category)

	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]struct{})
	recipients := []string{}
	for _, sub := range s.subscribers {
		if len(sub.Categories) > 0 && !slices.Contains(sub.Categories, category) {
			continue
		}
		if _, ok := seen[sub.Email]; ok {
			continue
		}
		seen[sub.Email] = struct{}{}
		recipients = append(recipients, sub.Email)
	}

	slices.Sort(recipients)
	return recipients
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
