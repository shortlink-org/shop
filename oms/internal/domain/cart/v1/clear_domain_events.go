package v1

// ClearDomainEvents clears the domain events list
// Should be called by application layer after publishing events
func (s *State) ClearDomainEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.domainEvents = s.domainEvents[:0]
}
