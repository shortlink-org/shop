package dto //nolint:dupl // Add and Remove DTOs intentionally share structure

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/rpcmeta"
	v2 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// AddRequestParams holds parsed parameters from AddRequest
type AddRequestParams struct {
	CustomerID uuid.UUID
	Items      []itemv1.Item
}

// AddRequestToDomain converts an AddRequest to domain parameters
func AddRequestToDomain(ctx context.Context, r *v2.AddRequest) (*AddRequestParams, error) {
	customerID, err := rpcmeta.CustomerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]itemv1.Item, 0, len(r.GetItems()))

	// parse items
	for i := range r.GetItems() {
		// string to uuid
		goodID, errParseItem := uuid.Parse(r.GetItems()[i].GetGoodId())
		if errParseItem != nil {
			return nil, ParseItemError{Err: errParseItem, item: r.GetItems()[i].GetGoodId()}
		}

		cartItem, err := itemv1.NewItem(goodID, r.GetItems()[i].GetQuantity())
		if err != nil {
			return nil, fmt.Errorf("invalid cart item %+v: %w", r.GetItems()[i], err)
		}

		items = append(items, cartItem)
	}

	return &AddRequestParams{
		CustomerID: customerID,
		Items:      items,
	}, nil
}
