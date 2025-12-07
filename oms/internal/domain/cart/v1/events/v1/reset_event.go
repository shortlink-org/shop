package v1

import (
	"time"

	"github.com/google/uuid"
)

// ResetEvent represents the domain event when the cart is reset
type ResetEvent struct {
	CustomerID uuid.UUID
	OccurredAt time.Time
}

func (e *ResetEvent) EventType() string {
	return "Reset"
}
