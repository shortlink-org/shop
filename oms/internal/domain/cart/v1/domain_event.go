package v1

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event that occurred in the aggregate.
type DomainEvent interface {
	EventType() string
}

// CartItemAddedEvent represents the domain event when an item is added to the cart
type CartItemAddedEvent struct {
	CustomerID uuid.UUID
	Item       CartItem
	OccurredAt time.Time
}

func (e *CartItemAddedEvent) EventType() string {
	return "CartItemAdded"
}

// CartItemRemovedEvent represents the domain event when an item is removed from the cart
type CartItemRemovedEvent struct {
	CustomerID uuid.UUID
	Item       CartItem
	OccurredAt time.Time
}

func (e *CartItemRemovedEvent) EventType() string {
	return "CartItemRemoved"
}

// CartResetEvent represents the domain event when the cart is reset
type CartResetEvent struct {
	CustomerID uuid.UUID
	OccurredAt time.Time
}

func (e *CartResetEvent) EventType() string {
	return "CartReset"
}



