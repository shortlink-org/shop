package v1

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/remove_items"
)

// Remove removes items from the cart
func (c *CartRPC) Remove(ctx context.Context, in *v1.RemoveRequest) (*emptypb.Empty, error) {
	params, err := dto.RemoveRequestToDomain(in)
	if err != nil {
		return nil, err
	}

	// Create command and execute handler
	cmd := remove_items.NewCommand(params.CustomerID, params.Items)
	if err := c.removeItemsHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
