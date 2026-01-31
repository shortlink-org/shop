package activities

import (
	"context"

	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	orderCancel "github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	orderGet "github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"
)

// Activities wraps order command/query handlers for Temporal activities.
// Activities are the bridge between Temporal workflows and application use cases.
// Temporal workflows must never access repositories directly - only through activities.
//
// Note: In the event-driven architecture, order creation happens before the workflow starts
// (CreateOrder command handler publishes event, which triggers the workflow).
// Activities are used for compensation (cancel) and queries during workflow execution.
type Activities struct {
	cancelHandler ports.CommandHandler[orderCancel.Command]
	getHandler    ports.QueryHandler[orderGet.Query, orderGet.Result]
}

// New creates a new Activities instance.
func New(
	cancelHandler ports.CommandHandler[orderCancel.Command],
	getHandler ports.QueryHandler[orderGet.Query, orderGet.Result],
) *Activities {
	return &Activities{
		cancelHandler: cancelHandler,
		getHandler:    getHandler,
	}
}

// CancelOrderRequest represents the request for CancelOrder activity.
type CancelOrderRequest struct {
	OrderID uuid.UUID
}

// CancelOrder cancels an order in the database.
// This is used for compensation in saga patterns.
func (a *Activities) CancelOrder(ctx context.Context, req CancelOrderRequest) error {
	cmd := orderCancel.NewCommand(req.OrderID)
	return a.cancelHandler.Handle(ctx, cmd)
}

// GetOrderRequest represents the request for GetOrder activity.
type GetOrderRequest struct {
	OrderID uuid.UUID
}

// GetOrderResponse represents the response from GetOrder activity.
type GetOrderResponse struct {
	Order *orderv1.OrderState
}

// GetOrder retrieves an order from the database.
func (a *Activities) GetOrder(ctx context.Context, req GetOrderRequest) (*GetOrderResponse, error) {
	query := orderGet.NewQuery(req.OrderID)
	order, err := a.getHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	return &GetOrderResponse{Order: order}, nil
}
