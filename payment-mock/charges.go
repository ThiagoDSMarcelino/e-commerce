package main

import "sync"

type ChargeStatus string

const (
	StatusPending    ChargeStatus = "PENDENTE"
	StatusProcessing ChargeStatus = "PROCESSANDO"
	StatusApproved   ChargeStatus = "APROVADO"
	StatusDeclined   ChargeStatus = "RECUSADO"
	StatusFailed     ChargeStatus = "FALHOU"
)

type Charge struct {
	OrderId    string
	WebhookUrl string
	Status     ChargeStatus
	Error      string
}

type ChargeStore struct {
	mu      sync.Mutex
	charges map[string]*Charge
}

func NewChargeStore() *ChargeStore {
	return &ChargeStore{charges: make(map[string]*Charge)}
}

func (s *ChargeStore) Create(orderId, webhookUrl string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.charges[orderId]; ok && c.Status != StatusPending {
		return false
	}

	s.charges[orderId] = &Charge{OrderId: orderId, WebhookUrl: webhookUrl, Status: StatusPending}

	return true
}

func (s *ChargeStore) Get(orderId string) (Charge, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.charges[orderId]
	if !ok {
		return Charge{}, false
	}

	return *c, true
}

func (s *ChargeStore) StartProcessing(orderId string) (Charge, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.charges[orderId]
	if !ok || c.Status != StatusPending {
		return Charge{}, false
	}

	c.Status = StatusProcessing
	c.Error = ""

	return *c, true
}

func (s *ChargeStore) Finish(orderId string, status ChargeStatus, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.charges[orderId]; ok {
		c.Status = status
		c.Error = errMsg
	}
}
