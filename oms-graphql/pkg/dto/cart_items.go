package dto

import (
	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

// CartItemInputsToOMS maps Connect item request to OMS cart items.
func CartItemInputsToOMS(input *servicepb.ItemRequest) []*cartmodel.CartItem {
	if input == nil {
		return nil
	}

	items := make([]*cartmodel.CartItem, 0, len(input.GetItems()))
	for _, item := range input.GetItems() {
		items = append(items, &cartmodel.CartItem{
			GoodId:   item.GetGoodId(),
			Quantity: item.GetQuantity(),
		})
	}

	return items
}
