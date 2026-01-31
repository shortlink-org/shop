package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// Add adds an item to the cart using the pattern: Load -> domain method -> Save
func (uc *UC) Add(ctx context.Context, customerID uuid.UUID, item itemv1.Item) error {
	// 1. Load aggregate (or create new if not found)
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			cart = v1.New(customerID)
		} else {
			return err
		}
	}

	// 2. Call domain method (business logic)
	if err := cart.AddItem(item); err != nil {
		return err
	}

	// 3. Save aggregate
	if err := uc.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Update index: add goods to customer's cart index
	uc.goodsIndex.AddGoodToCart(item.GetGoodId(), customerID)

	return nil
}

// AddItems adds multiple items to the cart.
func (uc *UC) AddItems(ctx context.Context, customerID uuid.UUID, items []itemv1.Item) error {
	// 1. Load aggregate (or create new if not found)
	cart, err := uc.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			cart = v1.New(customerID)
		} else {
			return err
		}
	}

	// 2. Call domain method for each item
	for _, item := range items {
		if err := cart.AddItem(item); err != nil {
			return err
		}
	}

	// 3. Save aggregate
	if err := uc.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Update index
	for _, item := range items {
		uc.goodsIndex.AddGoodToCart(item.GetGoodId(), customerID)
	}

	return nil
}
