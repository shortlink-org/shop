package dto

import (
	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	v2 "github.com/shortlink-org/shop/oms/internal/workers/cart/workflow/model/cart/v1"
)

// CartStateToCartEvent converts a list of items to a cart event.
func CartStateToCartEvent(in *v1.CartState, event v1.Event) v2.CartEvent {
	return v2.CartEvent{
		Event: event,
		Items: newWorkerCartItems(in.GetItems()),
	}
}

// newWorkerCartItems creates a new WorkerCartItems.
func newWorkerCartItems(in v1.CartItems) []*v2.WorkerCartItem {
	items := make([]*v2.WorkerCartItem, len(in))

	for i, item := range in {
		items[i] = &v2.WorkerCartItem{
			GoodId:   item.GetGoodId().String(),
			Quantity: item.GetQuantity(),
		}
	}

	return items
}
