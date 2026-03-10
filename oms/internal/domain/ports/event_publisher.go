package ports

import (
	"context"

	domainevents "github.com/shortlink-org/shop/oms/internal/domain/events"
)

// Event is an alias to the canonical domain event contract.
type Event = domainevents.Event

// EventPublisher publishes domain events (e.g. to outbox/Kafka via go-sdk/cqrs EventBus).
//
//nolint:iface // port interface used by usecases and DI
type EventPublisher interface {
	Publish(ctx context.Context, event any) error
}

// EventSubscriber subscribes to domain events.
