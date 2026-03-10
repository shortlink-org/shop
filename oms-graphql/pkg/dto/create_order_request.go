package dto

import (
	"github.com/google/uuid"

	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

// CreateOrderRequestFromInput maps Connect create order input to OMS CreateRequest.
// CustomerId is left empty; OMS fills it from request metadata (x-user-id set by Istio).
func CreateOrderRequestFromInput(input *servicepb.CreateOrderInput) (*ordermodel.CreateRequest, error) {
	if input == nil {
		return nil, InvalidArgument("input is required")
	}

	deliveryInfo, err := DeliveryInfoFromInput(input.GetDeliveryInfo())
	if err != nil {
		return nil, err
	}

	items := make([]*ordermodel.OrderItem, 0, len(input.GetItems()))
	for _, item := range input.GetItems() {
		items = append(items, &ordermodel.OrderItem{
			Id:       item.GetId(),
			Quantity: item.GetQuantity(),
			Price:    item.GetPrice(),
		})
	}

	return &ordermodel.CreateRequest{
		Order: &ordermodel.OrderState{
			Id:    uuid.NewString(),
			Items: items,
		},
		DeliveryInfo: deliveryInfo,
	}, nil
}
