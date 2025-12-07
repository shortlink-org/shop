package v1

import (
	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/events/v1"
)

// addDomainEvent adds a domain event to the aggregate's event list
func (s *State) addDomainEvent(event eventsv1.DomainEvent) {
	s.domainEvents = append(s.domainEvents, event)
}
