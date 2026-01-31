package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// Remove removes an item from the cart using the pattern: Load -> domain method -> Save
func (uc *UC) Remove(ctx context.Context, customerID uuid.UUID, item itemv1.Item) error {
	// 1. Load aggregate
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Cart doesn't exist, nothing to remove
			return nil
		}
		return err
	}

	// 2. Call domain method (business logic)
	if err := cart.RemoveItem(item); err != nil {
		return err
	}

	// 3. Save aggregate
	if err := uc.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Update index
	uc.goodsIndex.RemoveGoodFromCart(item.GetGoodId(), customerID)

	return nil
}

// RemoveItems removes multiple items from the cart.
func (uc *UC) RemoveItems(ctx context.Context, customerID uuid.UUID, items []itemv1.Item) error {
	// 1. Load aggregate
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil
		}
		return err
	}

	// 2. Call domain method for each item
	for _, item := range items {
		if err := cart.RemoveItem(item); err != nil {
			return err
		}
	}

	// 3. Save aggregate
	if err := uc.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Update index
	for _, item := range items {
		uc.goodsIndex.RemoveGoodFromCart(item.GetGoodId(), customerID)
	}

	return nil
}
