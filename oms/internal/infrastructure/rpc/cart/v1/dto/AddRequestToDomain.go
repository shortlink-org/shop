package dto

import (
	"fmt"

	"github.com/google/uuid"

	domain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	v2 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// AddRequestToDomain converts an AddRequest to a domain model
func AddRequestToDomain(r *v2.AddRequest) (*domain.CartState, error) {
	// string to uuid
	customerId, err := uuid.Parse(r.CustomerId)
	if err != nil {
		return nil, ErrInvalidCustomerId
	}

	// create a domain model
	item := domain.NewCartState(customerId)

	// add item to the cart
	for i := range r.GetItems() {
		// string to uuid
		goodId, errParseItem := uuid.Parse(r.Items[i].GoodId)
		if errParseItem != nil {
			return nil, ParseItemError{Err: errParseItem, item: r.Items[i].GoodId}
		}

		cartItem, err := domain.NewCartItem(goodId, r.Items[i].Quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid cart item %+v: %w", r.Items[i], err)
		}

		if err := item.AddItem(cartItem); err != nil {
			return nil, fmt.Errorf("failed to add cart item %+v: %w", r.Items[i], err)
		}
	}

	return item, nil
}
