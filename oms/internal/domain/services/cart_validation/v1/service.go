package v1

import (
	"context"

	"github.com/google/uuid"
)

// Service is a domain service that validates cart operations.
// This service contains business rules for cart validation that don't belong to a single entity.
type Service struct {
	// StockChecker is an interface for checking stock availability
	StockChecker StockChecker
}

// StockChecker defines the interface for checking stock availability
type StockChecker interface {
	// CheckStockAvailability checks if a good has sufficient stock
	CheckStockAvailability(ctx context.Context, goodId uuid.UUID, requestedQuantity int32) (bool, uint32, error)
}

// New creates a new validation Service
func New(stockChecker StockChecker) *Service {
	return &Service{
		StockChecker: stockChecker,
	}
}

