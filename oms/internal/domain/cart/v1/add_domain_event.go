package v1

import (
	domainevents "github.com/shortlink-org/shop/oms/internal/domain/events"
)

// addDomainEvent adds a domain event to the aggregate's event list
func (s *State) addDomainEvent(event domainevents.Event) {
	s.domainEvents = append(s.domainEvents, event)
}
