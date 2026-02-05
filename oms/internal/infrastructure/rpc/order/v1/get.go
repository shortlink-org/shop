package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"
)

func (o *OrderRPC) Get(ctx context.Context, in *v1.GetRequest) (*v1.GetResponse, error) {
	// parse order ID to UUID
	orderId, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, err
	}

	// Create query and execute handler
	query := get.NewQuery(orderId)

	orderState, err := o.getHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	return &v1.GetResponse{
		Order: dto.DomainToOrderState(orderState),
	}, nil
}
