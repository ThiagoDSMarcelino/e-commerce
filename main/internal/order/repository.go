package order

import (
	"slices"
	"sync"
)

type OrdersRepository struct {
	mu     sync.Mutex
	orders []Order
	nextId int
}

func NewOrdersRepository() *OrdersRepository {
	return &OrdersRepository{nextId: 1}
}

// func (s *OrdersRepository) NextId() string {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
//
// 	id := fmt.Sprintf("PED-%03d", s.nextId)
// 	s.nextId++
//
// 	return id
// }
//
// func (s *OrdersRepository) Add(order Order, status OrderStatus) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
//
// 	for i := range s.orders {
// 		if s.orders[i].Order.Id == order.Id {
// 			s.orders[i] = Order{Order: order, Status: status}
// 			return
// 		}
// 	}
//
// 	s.orders = append(s.orders, Order{Order: order, Status: status})
// }
//
// func (s *OrdersRepository) SetStatus(id string, status OrderStatus) bool {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
//
// 	for i := range s.orders {
// 		if s.orders[i].Order.Id == id {
// 			s.orders[i].Status = status
// 			return true
// 		}
// 	}
//
// 	return false
// }
//
// func (s *OrdersRepository) MarkCancelled(id string) (Order, bool) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
//
// 	for i := range s.orders {
// 		if s.orders[i].Order.Id != id {
// 			continue
// 		}
//
// 		if s.orders[i].Cancelled {
// 			return Order{}, false
// 		}
//
// 		s.orders[i].Cancelled = true
//
// 		return s.orders[i].Order, true
// 	}
//
// 	return Order{}, false
// }

func (s *OrdersRepository) GetAll() []Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.orders) == 0 {
		return []Order{}
	}

	return slices.Clone(s.orders)
}
