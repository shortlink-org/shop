package dto

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

// CartStateToService maps OMS cart state to Connect/GraphQL service response.
func CartStateToService(state *cartmodel.CartState) *servicepb.CartState {
	if state == nil {
		return nil
	}

	items := make([]*servicepb.CartItem, 0, len(state.GetItems()))
	for _, item := range state.GetItems() {
		items = append(items, &servicepb.CartItem{
			GoodId:   wrapperspb.String(item.GetGoodId()),
			Quantity: wrapperspb.Int32(item.GetQuantity()),
		})
	}

	return &servicepb.CartState{
		CartId: wrapperspb.String(state.GetCartId()),
		Items: &servicepb.ListOfCartItem{
			List: &servicepb.ListOfCartItem_List{
				Items: items,
			},
		},
	}
}
