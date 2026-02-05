package dto

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	v3 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
)

var (
	errInvalidCustomerID = errors.New("invalid customer id")
	errInvalidOrderID    = errors.New("invalid order id")
	errInvalidItemID     = errors.New("invalid item id")
)

// OrderStateToDomain converts a v3.OrderState to a v1.OrderState using the OrderStateBuilder
func OrderStateToDomain(in *v3.OrderState) (*v1.OrderState, error) {
	// Parse the customer ID
	customerID, err := uuid.Parse(in.GetCustomerId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidCustomerID, err)
	}

	// Initialize the builder with the customer ID
	builder := v1.NewOrderStateBuilder(customerID)

	// Set the order ID
	orderID, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidOrderID, err)
	}

	builder.SetId(orderID)

	// Add items to the order
	for _, item := range in.GetItems() {
		goodID, err := uuid.Parse(item.GetId())
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errInvalidItemID, err)
		}

		price := decimal.NewFromFloat(item.GetPrice())
		builder.AddItem(goodID, item.GetQuantity(), price)
	}

	// Set the status by replaying events (preserves FSM invariants)
	// Domain layer no longer depends on context.Context
	builder.SetStatus(in.GetStatus())

	// Build the OrderState
	orderState, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return orderState, nil
}
