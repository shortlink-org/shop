package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// ValidateRemoveItems validates if items can be removed from a cart.
func (s *Service) ValidateRemoveItems(
	ctx context.Context,
	cart *v1.State,
	items itemsv1.Items,
) Result {
	result := Result{
		Valid:    true,
		Errors:   make([]Error, 0),
		Warnings: make([]Warning, 0),
	}

	cartItems := cart.GetItems()

	cartItemMap := make(map[uuid.UUID]itemv1.Item)
	for _, item := range cartItems {
		cartItemMap[item.GetGoodId()] = item
	}

	for _, item := range items {
		cartItem, exists := cartItemMap[item.GetGoodId()]
		if !exists {
			result.Warnings = append(result.Warnings, Warning{
				GoodID:  item.GetGoodId(),
				Message: "Item not found in cart",
				Code:    "ITEM_NOT_IN_CART",
			})

			continue
		}

		// Check if removing more than available
		if item.GetQuantity() > cartItem.GetQuantity() {
			result.Warnings = append(result.Warnings, Warning{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Removing more items (%d) than in cart (%d)", item.GetQuantity(), cartItem.GetQuantity()),
				Code:    "REMOVE_EXCEEDS_CART",
			})
		}
	}

	return result
}
