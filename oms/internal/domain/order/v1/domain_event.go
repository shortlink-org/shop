package v1

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event that occurred in the aggregate.
// This is a marker interface - specific event types should implement it.
// Domain layer doesn't depend on proto implementations - application layer handles conversion.
type DomainEvent interface {
	// EventType returns the type name of the event (e.g., "OrderCreated", "OrderCancelled")
	EventType() string
}

// OrderCreatedEvent represents the domain event when an order is created
type OrderCreatedEvent struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	Items      Items
	Status     OrderStatus
	OccurredAt time.Time
}

func (e *OrderCreatedEvent) EventType() string {
	return "OrderCreated"
}

// OrderCancelledEvent represents the domain event when an order is cancelled
type OrderCancelledEvent struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	Status     OrderStatus
	Reason     string
	OccurredAt time.Time
}

func (e *OrderCancelledEvent) EventType() string {
	return "OrderCancelled"
}

// OrderCompletedEvent represents the domain event when an order is completed
type OrderCompletedEvent struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	Status     OrderStatus
	OccurredAt time.Time
}

func (e *OrderCompletedEvent) EventType() string {
	return "OrderCompleted"
}

