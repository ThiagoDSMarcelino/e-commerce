package registrations

import "context"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) RegisterEmail(ctx context.Context, event RegisterPromotionEvent) error {
	return nil
}

func (s *Service) UnregisterEmail(ctx context.Context, event UnregisterPromotionEvent) error {
	return nil
}

func (s *Service) GetRegistredEmails() {

}
