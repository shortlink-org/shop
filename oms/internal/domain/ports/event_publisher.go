package ports

import "context"

// Event is a marker interface for domain events.
// Uses EventType() to align with existing domain event pattern.
type Event interface {
	EventType() string
}

// EventPublisher publishes domain events to subscribers.
type EventPublisher interface {
	// Publish publishes an event to all registered subscribers.
	Publish(ctx context.Context, event Event) error
}

// EventSubscriber subscribes to domain events.
type EventSubscriber interface {
	// Subscribe registers a handler for events of a specific type.
	Subscribe(eventType string, handler func(ctx context.Context, event Event) error)
}
