package v1

import (
	"sync"

	"github.com/google/uuid"

	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/events/v1"
	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// State represents the cart state.
type State struct {
	mu sync.Mutex

	// items is the cart items
	items itemsv1.Items
	// customerId is the customer ID
	customerId uuid.UUID
	// domainEvents stores domain events that occurred during aggregate operations
	domainEvents []eventsv1.DomainEvent
}

// New creates a new cart state.
func New(customerId uuid.UUID) *State {
	return &State{
		items:        make(itemsv1.Items, 0),
		customerId:   customerId,
		domainEvents: make([]eventsv1.DomainEvent, 0),
	}
}

