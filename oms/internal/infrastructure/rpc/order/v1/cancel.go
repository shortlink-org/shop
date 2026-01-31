package v1

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
)

func (o *OrderRPC) Cancel(ctx context.Context, in *v1.CancelRequest) (*emptypb.Empty, error) {
	// parse order ID to UUID
	orderId, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, err
	}

	// Create command and execute handler
	cmd := cancel.NewCommand(orderId)
	if err := o.cancelHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
