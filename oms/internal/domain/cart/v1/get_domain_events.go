package v1

import (
	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/events/v1"
)

// GetDomainEvents returns all domain events that occurred during aggregate operations
// Application layer should call this after aggregate operations to publish events
// and then call ClearDomainEvents() to reset the list
func (s *State) GetDomainEvents() []eventsv1.DomainEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy to prevent external modification
	eventsCopy := make([]eventsv1.DomainEvent, len(s.domainEvents))
	copy(eventsCopy, s.domainEvents)
	return eventsCopy
}

