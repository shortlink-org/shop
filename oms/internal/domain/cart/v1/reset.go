package v1

import (
	"time"

	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/events/v1"
	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// Reset resets the cart.
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(itemsv1.Items, 0)
	// Generate domain event for cart reset
	s.addDomainEvent(&eventsv1.ResetEvent{
		CustomerID: s.customerId,
		OccurredAt: time.Now(),
	})
}

