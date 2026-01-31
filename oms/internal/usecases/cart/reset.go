package cart

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
)

// Reset resets the cart using the pattern: Load -> domain method -> Save
func (uc *UC) Reset(ctx context.Context, customerID uuid.UUID) error {
	// Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = uc.uow.Rollback(ctx) }()

	// 1. Load aggregate
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Cart doesn't exist, nothing to reset
			return nil
		}
		return err
	}

	// 2. Call domain method (business logic)
	cart.Reset()

	// 3. Save aggregate
	if err := uc.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Commit transaction
	if err := uc.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear index for this customer
	if err := uc.goodsIndex.ClearCart(ctx, customerID); err != nil {
		// Log but don't fail - index is eventually consistent
		uc.log.Warn("failed to clear cart goods index", slog.Any("error", err))
	}

	return nil
}
