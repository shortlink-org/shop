package v1

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/grpcerr"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/rpcmeta"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/reset"
)

// Reset resets the cart
func (c *CartRPC) Reset(ctx context.Context, in *v1.ResetRequest) (*emptypb.Empty, error) {
	customerID, err := rpcmeta.CustomerIDFromContext(ctx)
	if err != nil {
		return nil, grpcerr.ToStatus(ctx, c.log, "Cart.Reset", err)
	}

	cmd := reset.NewCommand(customerID)
	if err := c.resetHandler.Handle(ctx, cmd); err != nil {
		return nil, grpcerr.ToStatus(ctx, c.log, "Cart.Reset", err)
	}

	return &emptypb.Empty{}, nil
}
