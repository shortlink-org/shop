package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
)

// Get retrieves the cart for a customer.
func (uc *UC) Get(ctx context.Context, customerID uuid.UUID) (*v1.State, error) {
	// Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = uc.uow.Rollback(ctx) }()

	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Return empty cart if not found
			return v1.New(customerID), nil
		}
		return nil, err
	}

	// Commit transaction (read-only, but still needs to close tx)
	if err := uc.uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return cart, nil
}
