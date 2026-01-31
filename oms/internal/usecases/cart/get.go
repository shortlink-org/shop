package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
)

// Get retrieves the cart for a customer.
func (uc *UC) Get(ctx context.Context, customerID uuid.UUID) (*v1.State, error) {
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Return empty cart if not found
			return v1.New(customerID), nil
		}
		return nil, err
	}

	return cart, nil
}
