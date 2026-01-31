package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/query/get"
)

func (c *CartRPC) Get(ctx context.Context, in *v1.GetRequest) (*v1.GetResponse, error) {
	// customerId to uuid
	customerId, err := uuid.Parse(in.CustomerId)
	if err != nil {
		return nil, err
	}

	// Create query and execute handler
	query := get.NewQuery(customerId)
	response, err := c.getHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	return dto.GetResponseFromDomain(response), nil
}
