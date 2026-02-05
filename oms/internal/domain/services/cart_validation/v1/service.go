package v1

// Service is a domain service that validates cart operations.
// It contains business rules for cart validation that don't belong to a single entity.
// It has no I/O dependencies; the use case must supply pre-fetched data (e.g. stock) when needed.
type Service struct{}

// New creates a new validation Service (no dependencies).
func New() *Service {
	return &Service{}
}
