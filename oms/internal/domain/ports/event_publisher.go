package ports

import "context"

// Event is a marker interface for domain events.
// Uses EventType() to align with existing domain event pattern.
type Event interface {
	EventType() string
}

// EventPublisher publishes domain events (e.g. to outbox/Kafka via go-sdk/cqrs EventBus).
type EventPublisher interface {
	Publish(ctx context.Context, event any) error
}

// EventSubscriber subscribes to domain events.
