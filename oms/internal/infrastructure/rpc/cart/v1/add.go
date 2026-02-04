package v1

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shortlink-org/shop/oms/internal/domain"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/grpcerr"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/add_items"
)

// Add adds items to the cart
func (c *CartRPC) Add(ctx context.Context, in *v1.AddRequest) (*emptypb.Empty, error) {
	params, err := dto.AddRequestToDomain(in)
	if err != nil {
		return nil, grpcerr.ToStatus(ctx, c.log, "Cart.Add", domain.WrapValidation("AddRequestToDomain", err))
	}

	cmd := add_items.NewCommand(params.CustomerID, params.Items)
	if err := c.addItemsHandler.Handle(ctx, cmd); err != nil {
		return nil, grpcerr.ToStatus(ctx, c.log, "Cart.Add", err)
	}

	return &emptypb.Empty{}, nil
}
