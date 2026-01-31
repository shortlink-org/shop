package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
)

// Reset resets the cart using the pattern: Load -> domain method -> Save
func (uc *UC) Reset(ctx context.Context, customerID uuid.UUID) error {
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

	// Clear index for this customer
	uc.goodsIndex.ClearCart(customerID)

	return nil
}
