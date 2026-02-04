package v1

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shortlink-org/shop/oms/internal/domain"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/grpcerr"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/reset"
)

// Reset resets the cart
func (c *CartRPC) Reset(ctx context.Context, in *v1.ResetRequest) (*emptypb.Empty, error) {
	customerId, err := uuid.Parse(in.CustomerId)
	if err != nil {
		return nil, grpcerr.ToStatus(ctx, c.log, "Cart.Reset", domain.WrapValidation("customer_id", err))
	}

	cmd := reset.NewCommand(customerId)
	if err := c.resetHandler.Handle(ctx, cmd); err != nil {
		return nil, grpcerr.ToStatus(ctx, c.log, "Cart.Reset", err)
	}

	return &emptypb.Empty{}, nil
}
