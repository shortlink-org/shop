package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/dto"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/command/create_order_from_cart"
)

// Checkout creates an order from customer's cart.
func (o *OrderRPC) Checkout(ctx context.Context, in *v1.CheckoutRequest) (*v1.CheckoutResponse, error) {
	// Parse customer ID to UUID
	customerID, err := uuid.Parse(in.GetCustomerId())
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	// Convert proto delivery info to domain (can be nil for self-pickup)
	deliveryInfo := dto.ProtoDeliveryInfoToDomain(in.GetDeliveryInfo())

	// Create command and execute handler
	cmd := create_order_from_cart.NewCommand(customerID, deliveryInfo)

	result, err := o.checkoutHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &v1.CheckoutResponse{
		OrderId:       result.Order.GetOrderID().String(),
		Subtotal:      result.Subtotal.InexactFloat64(),
		TotalDiscount: result.TotalDiscount.InexactFloat64(),
		TotalTax:      result.TotalTax.InexactFloat64(),
		FinalPrice:    result.FinalPrice.InexactFloat64(),
	}, nil
}
