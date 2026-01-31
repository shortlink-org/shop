package dto

import (
	"fmt"

	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	v2 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// AddRequestParams holds parsed parameters from AddRequest
type AddRequestParams struct {
	CustomerID uuid.UUID
	Items      []itemv1.Item
}

// AddRequestToDomain converts an AddRequest to domain parameters
func AddRequestToDomain(r *v2.AddRequest) (*AddRequestParams, error) {
	// string to uuid
	customerID, err := uuid.Parse(r.CustomerId)
	if err != nil {
		return nil, ErrInvalidCustomerId
	}

	items := make([]itemv1.Item, 0, len(r.GetItems()))

	// parse items
	for i := range r.GetItems() {
		// string to uuid
		goodID, errParseItem := uuid.Parse(r.Items[i].GoodId)
		if errParseItem != nil {
			return nil, ParseItemError{Err: errParseItem, item: r.Items[i].GoodId}
		}

		cartItem, err := itemv1.NewItem(goodID, r.Items[i].Quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid cart item %+v: %w", r.Items[i], err)
		}

		items = append(items, cartItem)
	}

	return &AddRequestParams{
		CustomerID: customerID,
		Items:      items,
	}, nil
}
