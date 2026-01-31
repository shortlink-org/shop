package v1

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// Add adds items to the cart
func (c *CartRPC) Add(ctx context.Context, in *v1.AddRequest) (*emptypb.Empty, error) {
	params, err := dto.AddRequestToDomain(in)
	if err != nil {
		return nil, err
	}

	// Add items using the new UseCase signature
	if err := c.cartService.AddItems(ctx, params.CustomerID, params.Items); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
