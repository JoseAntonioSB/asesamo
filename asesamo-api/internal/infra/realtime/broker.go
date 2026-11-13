package realtime

import (
	"sync"
	"time"
)

// Event represents a realtime push event broadcasted to connected clients.
type Event struct {
	Type      string                 `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// Broker defines the contract for multi-tenant event pub/sub.
type Broker interface {
	Publish(event Event)
	Subscribe(tenantID string) (<-chan Event, func())
}

// InMemoryEventBroker is a high-performance, thread-safe in-memory event broker.
type InMemoryEventBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Event]struct{}
	bufSize     int
}

// NewInMemoryEventBroker creates an initialized broker with a default channel buffer per subscriber.
func NewInMemoryEventBroker(bufSize int) *InMemoryEventBroker {
	if bufSize <= 0 {
		bufSize = 32
	}
	return &InMemoryEventBroker{
		subscribers: make(map[string]map[chan Event]struct{}),
		bufSize:     bufSize,
	}
}

// Subscribe registers a new subscriber channel for a given tenantID and returns an unsubscribe cleanup function.
func (b *InMemoryEventBroker) Subscribe(tenantID string) (<-chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, b.bufSize)
	if _, ok := b.subscribers[tenantID]; !ok {
		b.subscribers[tenantID] = make(map[chan Event]struct{})
	}
	b.subscribers[tenantID][ch] = struct{}{}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			if tenantSubs, ok := b.subscribers[tenantID]; ok {
				delete(tenantSubs, ch)
				if len(tenantSubs) == 0 {
					delete(b.subscribers, tenantID)
				}
			}
			close(ch)
		})
	}

	return ch, unsubscribe
}

// Publish broadcasts an event to all active subscribers for the event's TenantID.
// Slow subscribers with full channel buffers are dropped gracefully without blocking other clients.
func (b *InMemoryEventBroker) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tenantSubs, ok := b.subscribers[event.TenantID]
	if !ok || len(tenantSubs) == 0 {
		return
	}

	for ch := range tenantSubs {
		select {
		case ch <- event:
		default:
			// Buffer full, drop or skip to avoid blocking the broker
		}
	}
}
