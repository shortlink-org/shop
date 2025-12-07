package cart

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// HandleStockChange handles stock change events and removes out-of-stock items from carts
func (uc *UC) HandleStockChange(ctx context.Context, goodId uuid.UUID, newQuantity uint32) error {
	// If stock is not zero, no action needed
	if newQuantity > 0 {
		return nil
	}

	uc.log.Info("Stock depleted for good", "good_id", goodId.String())

	// Get all customers that have this good in their cart using the index
	customerIds := uc.goodsIndex.GetCustomersWithGood(goodId)
	if len(customerIds) == 0 {
		uc.log.Info("No carts found with the out-of-stock item", "good_id", goodId.String())
		return nil
	}

	// Process each cart
	for _, customerId := range customerIds {
		// Get cart state to get the quantity of the item
		cartState, err := uc.Get(ctx, customerId)
		if err != nil {
			uc.log.Warn("Failed to get cart state",
				"customer_id", customerId.String(),
				"good_id", goodId.String(),
				"error", err)
			continue
		}

		// Find the item and its quantity
		var itemQuantity int32
		found := false
		for _, item := range cartState.GetItems() {
			if item.GetGoodId() == goodId {
				itemQuantity = item.GetQuantity()
				found = true
				break
			}
		}

		if !found {
			// Item was already removed, clean up index
			uc.goodsIndex.RemoveGoodFromCart(goodId, customerId)
			continue
		}

		// Remove the item from cart
		uc.log.Info("Removing out-of-stock item from cart",
			"customer_id", customerId.String(),
			"good_id", goodId.String(),
			"quantity", itemQuantity)

		removeRequest := domain.New(customerId)
		cartItem, err := itemv1.NewItem(goodId, itemQuantity)
		if err != nil {
			uc.log.Warn("Failed to construct cart item for removal",
				"customer_id", customerId.String(),
				"good_id", goodId.String(),
				"error", err)
			continue
		}

		if err := removeRequest.AddItem(cartItem); err != nil {
			uc.log.Warn("Failed to stage cart item removal",
				"customer_id", customerId.String(),
				"good_id", goodId.String(),
				"error", err)
			continue
		}

		err = uc.Remove(ctx, removeRequest)
		if err != nil {
			uc.log.Warn("Failed to remove item from cart",
				"customer_id", customerId.String(),
				"good_id", goodId.String(),
				"error", err)
			continue
		}

		// Remove from index (already done in Remove, but ensure it's done)
		uc.goodsIndex.RemoveGoodFromCart(goodId, customerId)

		// Send websocket notification to UI
		if uc.notifier != nil {
			if err := uc.notifier.NotifyStockDepleted(customerId, goodId); err != nil {
				uc.log.Warn("Failed to send websocket notification",
					"customer_id", customerId.String(),
					"good_id", goodId.String(),
					"error", err)
			}
		}
	}

	return nil
}
