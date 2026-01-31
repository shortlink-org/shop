package events

import (
	"context"
	"sync"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// InMemoryPublisher is an in-memory event publisher that dispatches events to subscribers.
// It implements both EventPublisher and EventSubscriber interfaces.
type InMemoryPublisher struct {
	mu       sync.RWMutex
	handlers map[string][]func(ctx context.Context, event ports.Event) error
}

// NewInMemoryPublisher creates a new in-memory event publisher.
func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		handlers: make(map[string][]func(ctx context.Context, event ports.Event) error),
	}
}

// Publish publishes an event to all registered subscribers.
// Returns the first error encountered, but continues to call all handlers.
func (p *InMemoryPublisher) Publish(ctx context.Context, event ports.Event) error {
	p.mu.RLock()
	handlers := p.handlers[event.EventType()]
	p.mu.RUnlock()

	var firstErr error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Subscribe registers a handler for events of a specific type.
func (p *InMemoryPublisher) Subscribe(eventType string, handler func(ctx context.Context, event ports.Event) error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers[eventType] = append(p.handlers[eventType], handler)
}

// SubscribeTyped is a helper to subscribe with a typed handler.
// It wraps the typed handler to match the generic signature.
func SubscribeTyped[E ports.Event](p *InMemoryPublisher, handler func(ctx context.Context, event E) error) {
	var zero E
	eventType := zero.EventType()

	p.Subscribe(eventType, func(ctx context.Context, event ports.Event) error {
		typedEvent, ok := event.(E)
		if !ok {
			return nil // Skip if type doesn't match
		}
		return handler(ctx, typedEvent)
	})
}

// Ensure InMemoryPublisher implements both interfaces.
var (
	_ ports.EventPublisher  = (*InMemoryPublisher)(nil)
	_ ports.EventSubscriber = (*InMemoryPublisher)(nil)
)
