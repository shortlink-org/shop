package v1

import (
	"sync"

	"github.com/google/uuid"

	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	domainevents "github.com/shortlink-org/shop/oms/internal/domain/events"
)

// State represents the cart state.
type State struct {
	mu sync.Mutex

	// items is the cart items
	items itemsv1.Items
	// customerId is the customer ID
	customerId uuid.UUID
	// version is used for optimistic concurrency control
	version int
	// domainEvents stores domain events that occurred during aggregate operations
	domainEvents []domainevents.Event
}

// New creates a new cart state.
func New(customerId uuid.UUID) *State {
	return &State{
		items:        make(itemsv1.Items, 0),
		customerId:   customerId,
		version:      0,
		domainEvents: make([]domainevents.Event, 0),
	}
}

// Reconstitute creates a cart state from persisted data.
// This is used by the repository to rebuild the aggregate from the database.
// It bypasses validation since the data is already validated when it was saved.
func Reconstitute(customerId uuid.UUID, items itemsv1.Items, version int) *State {
	return &State{
		items:        items,
		customerId:   customerId,
		version:      version,
		domainEvents: make([]domainevents.Event, 0),
	}
}

// GetVersion returns the current version for optimistic concurrency control.
func (s *State) GetVersion() int {
	return s.version
}
