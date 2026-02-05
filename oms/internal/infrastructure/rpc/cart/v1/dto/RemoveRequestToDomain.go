package dto

import (
	"fmt"

	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// RemoveRequestParams holds parsed parameters from RemoveRequest
type RemoveRequestParams struct {
	CustomerID uuid.UUID
	Items      []itemv1.Item
}

// RemoveRequestToDomain converts a RemoveRequest to domain parameters
func RemoveRequestToDomain(r *v1.RemoveRequest) (*RemoveRequestParams, error) {
	// string to uuid
	customerID, err := uuid.Parse(r.GetCustomerId())
	if err != nil {
		return nil, ErrInvalidCustomerId
	}

	items := make([]itemv1.Item, 0, len(r.GetItems()))

	// parse items
	for i := range r.GetItems() {
		// string to uuid
		goodID, errParseItem := uuid.Parse(r.GetItems()[i].GetGoodId())
		if errParseItem != nil {
			return nil, ParseItemError{Err: errParseItem, item: r.GetItems()[i].GetGoodId()}
		}

		// create CartItem
		item, err := itemv1.NewItem(goodID, r.GetItems()[i].GetQuantity())
		if err != nil {
			return nil, fmt.Errorf("invalid cart item %+v: %w", r.GetItems()[i], err)
		}

		items = append(items, item)
	}

	return &RemoveRequestParams{
		CustomerID: customerID,
		Items:      items,
	}, nil
}
