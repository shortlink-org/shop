package ports

import (
	"context"

	domainevents "github.com/shortlink-org/shop/oms/internal/domain/events"
)

// Event is an alias to the canonical domain event contract.
type Event = domainevents.Event

// EventPublisher publishes domain events (e.g. to outbox/Kafka via go-sdk/cqrs EventBus).
type EventPublisher interface {
	Publish(ctx context.Context, event any) error
}

// EventSubscriber subscribes to domain events.
type EventSubscriber interface {
	Subscribe(eventType string, handler func(ctx context.Context, event Event) error)
}
