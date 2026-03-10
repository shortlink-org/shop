package dto

import (
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// OrderStateToService maps OMS order state to Connect/GraphQL service response.
func OrderStateToService(order *ordermodel.OrderState) *servicepb.OrderState {
	if order == nil {
		return nil
	}

	items := make([]*servicepb.OrderItem, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, &servicepb.OrderItem{
			Id:       wrapperspb.String(item.GetId()),
			Quantity: wrapperspb.Int32(item.GetQuantity()),
			Price:    wrapperspb.Double(item.GetPrice()),
		})
	}

	return &servicepb.OrderState{
		Id: wrapperspb.String(order.GetId()),
		Items: &servicepb.ListOfOrderItem{
			List: &servicepb.ListOfOrderItem_List{
				Items: items,
			},
		},
		Status:       wrapperspb.String(order.GetStatus().String()),
		DeliveryInfo: DeliveryInfoToService(order.GetDeliveryInfo()),
	}
}
