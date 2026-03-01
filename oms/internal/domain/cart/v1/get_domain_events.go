package v1

import (
	domainevents "github.com/shortlink-org/shop/oms/internal/domain/events"
)

// GetDomainEvents returns all domain events that occurred during aggregate operations
// Application layer should call this after aggregate operations to publish events
// and then call ClearDomainEvents() to reset the list
func (s *State) GetDomainEvents() []domainevents.Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy to prevent external modification
	eventsCopy := make([]domainevents.Event, len(s.domainEvents))
	copy(eventsCopy, s.domainEvents)

	return eventsCopy
}
