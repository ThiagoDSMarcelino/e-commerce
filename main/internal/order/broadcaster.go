package orders

import (
	"sync"
	"time"
	"uuid"
)

const subscriberBuffer = 16

type StatusEvent struct {
	OrderID    uuid.UUID `json:"orderId"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updatedAt"`
	PaymentURL string    `json:"paymentUrl,omitempty"`
}

type Broadcaster struct {
	mu   sync.RWMutex
	subs map[int64]map[chan StatusEvent]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subs: make(map[int64]map[chan StatusEvent]struct{})}
}

func (b *Broadcaster) Subscribe(clientID int64) (<-chan StatusEvent, func()) {
	ch := make(chan StatusEvent, subscriberBuffer)

	b.mu.Lock()
	if b.subs[clientID] == nil {
		b.subs[clientID] = make(map[chan StatusEvent]struct{})
	}
	b.subs[clientID][ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs[clientID], ch)
			if len(b.subs[clientID]) == 0 {
				delete(b.subs, clientID)
			}
			b.mu.Unlock()
			close(ch)
		})
	}

	return ch, unsubscribe
}

func (b *Broadcaster) Publish(clientID int64, event StatusEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subs[clientID] {
		select {
		case ch <- event:
		default:
		}
	}
}
