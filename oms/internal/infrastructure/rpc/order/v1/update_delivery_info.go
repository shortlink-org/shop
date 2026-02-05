package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/command/update_delivery_info"
)

var errDeliveryInfoRequired = errors.New("delivery info is required")

func (o *OrderRPC) UpdateDeliveryInfo(ctx context.Context, in *v1.UpdateDeliveryInfoRequest) (*emptypb.Empty, error) {
	// Parse order ID to UUID
	orderID, err := uuid.Parse(in.GetOrderId())
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	// Convert proto delivery info to domain
	deliveryInfo := dto.ProtoDeliveryInfoToDomain(in.GetDeliveryInfo())
	if deliveryInfo == nil {
		return nil, errDeliveryInfoRequired
	}

	// Create command and execute handler
	cmd := update_delivery_info.NewCommand(orderID, *deliveryInfo)
	if err := o.updateDeliveryInfoHandler.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
