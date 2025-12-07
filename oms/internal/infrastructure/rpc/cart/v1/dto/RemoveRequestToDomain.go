package dto

import (
	"fmt"

	"github.com/google/uuid"

	domain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// RemoveRequestToDomain converts a RemoveRequest to a domain model
func RemoveRequestToDomain(r *v1.RemoveRequest) (*domain.State, error) {
	// string to uuid
	customerId, err := uuid.Parse(r.CustomerId)
	if err != nil {
		return nil, ErrInvalidCustomerId
	}

	// create a domain model
	state := domain.New(customerId)

	// remove items from the cart
	for i := range r.GetItems() {
		// string to uuid
		goodId, errParseItem := uuid.Parse(r.Items[i].GoodId)
		if errParseItem != nil {
			return nil, ParseItemError{Err: errParseItem, item: r.Items[i].GoodId}
		}

		// create CartItem and remove it from the state
		item, err := itemv1.NewItem(goodId, r.Items[i].Quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid cart item %+v: %w", r.Items[i], err)
		}

		if err := state.AddItem(item); err != nil {
			return nil, fmt.Errorf("failed to stage cart item removal %+v: %w", r.Items[i], err)
		}
	}

	return state, nil
}
