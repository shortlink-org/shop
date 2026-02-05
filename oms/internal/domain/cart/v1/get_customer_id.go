package v1

import (
	"github.com/google/uuid"
)

// GetCustomerId returns the value of the customerId field.
func (s *State) GetCustomerId() uuid.UUID {
	return s.customerId
}
