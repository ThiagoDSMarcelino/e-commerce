package main

import (
	"fmt"
	"slices"
	"sync"
)

type OrderStatus string

const (
	StatusCreated       OrderStatus = "Criado"
	StatusStockOk       OrderStatus = "Estoque confirmado"
	StatusOutOfStock    OrderStatus = "Estoque indisponível"
	StatusPaymentOk     OrderStatus = "Pagamento aprovado"
	StatusPaymentFailed OrderStatus = "Pagamento recusado"
	StatusShipped       OrderStatus = "Enviado"
	StatusCancelled     OrderStatus = "Excluído"
)

type StoredOrder struct {
	Order     Order
	Status    OrderStatus
	Cancelled bool
}

type OrderStore struct {
	mu     sync.Mutex
	orders []StoredOrder
	nextId int
}

var orders = NewOrderStore()

func NewOrderStore() *OrderStore {
	return &OrderStore{nextId: 1}
}

func (s *OrderStore) NextId() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("PED-%03d", s.nextId)
	s.nextId++

	return id
}

func (s *OrderStore) Add(order Order, status OrderStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orders {
		if s.orders[i].Order.Id == order.Id {
			s.orders[i] = StoredOrder{Order: order, Status: status}
			return
		}
	}

	s.orders = append(s.orders, StoredOrder{Order: order, Status: status})
}

func (s *OrderStore) SetStatus(id string, status OrderStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orders {
		if s.orders[i].Order.Id == id {
			s.orders[i].Status = status
			return true
		}
	}

	return false
}

func (s *OrderStore) MarkCancelled(id string) (Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.orders {
		if s.orders[i].Order.Id != id {
			continue
		}

		if s.orders[i].Cancelled {
			return Order{}, false
		}

		s.orders[i].Cancelled = true

		return s.orders[i].Order, true
	}

	return Order{}, false
}

func (s *OrderStore) List() []StoredOrder {
	s.mu.Lock()
	defer s.mu.Unlock()

	return slices.Clone(s.orders)
}
