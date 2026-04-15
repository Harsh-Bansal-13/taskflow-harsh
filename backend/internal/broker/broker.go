package broker

import (
	"sync"

	"github.com/google/uuid"
)

// EventBroker is a simple in-memory pub/sub for per-project task events.
// Subscribers receive a signal (struct{}) whenever a task mutation occurs.
type EventBroker struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[chan struct{}]struct{}
}

func New() *EventBroker {
	return &EventBroker{
		clients: make(map[uuid.UUID]map[chan struct{}]struct{}),
	}
}

// Subscribe returns a buffered channel that receives a signal on task mutation.
func (b *EventBroker) Subscribe(projectID uuid.UUID) chan struct{} {
	ch := make(chan struct{}, 4)
	b.mu.Lock()
	if b.clients[projectID] == nil {
		b.clients[projectID] = make(map[chan struct{}]struct{})
	}
	b.clients[projectID][ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a channel and cleans up the project entry if empty.
func (b *EventBroker) Unsubscribe(projectID uuid.UUID, ch chan struct{}) {
	b.mu.Lock()
	if clients, ok := b.clients[projectID]; ok {
		delete(clients, ch)
		if len(clients) == 0 {
			delete(b.clients, projectID)
		}
	}
	b.mu.Unlock()
}

// Publish signals all subscribers for the given project. Non-blocking.
func (b *EventBroker) Publish(projectID uuid.UUID) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients[projectID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
