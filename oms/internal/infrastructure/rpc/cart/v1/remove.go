package v1

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

// Remove removes items from the cart
func (c *CartRPC) Remove(ctx context.Context, in *v1.RemoveRequest) (*emptypb.Empty, error) {
	params, err := dto.RemoveRequestToDomain(in)
	if err != nil {
		return nil, err
	}

	// Remove items using the new UseCase signature
	if err := c.cartService.RemoveItems(ctx, params.CustomerID, params.Items); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
