package v1

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/command/reset"
)

// Reset resets the cart
func (c *CartRPC) Reset(ctx context.Context, in *v1.ResetRequest) (*emptypb.Empty, error) {
	// customerId to uuid
	customerId, err := uuid.Parse(in.CustomerId)
	if err != nil {
		return nil, err
	}

	// Create command and execute handler
	cmd := reset.NewCommand(customerId)
	if err := c.resetHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
