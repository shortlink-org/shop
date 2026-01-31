package activities

import (
	"context"

	"github.com/google/uuid"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	orderuc "github.com/shortlink-org/shop/oms/internal/usecases/order"
)

// Activities wraps order use case methods for Temporal activities.
// Activities are the bridge between Temporal workflows and application use cases.
// Temporal workflows must never access repositories directly - only through activities.
type Activities struct {
	orderUC *orderuc.UC
}

// New creates a new Activities instance.
func New(orderUC *orderuc.UC) *Activities {
	return &Activities{
		orderUC: orderUC,
	}
}

// CreateOrderRequest represents the request for CreateOrder activity.
type CreateOrderRequest struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
	Items      orderv1.Items
}

// CreateOrder creates an order in the database.
// This is called by the workflow after stock reservation.
func (a *Activities) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
	return a.orderUC.CreateInDB(ctx, req.OrderID, req.CustomerID, req.Items)
}

// CancelOrderRequest represents the request for CancelOrder activity.
type CancelOrderRequest struct {
	OrderID uuid.UUID
}

// CancelOrder cancels an order in the database.
// This is used for compensation in saga patterns.
func (a *Activities) CancelOrder(ctx context.Context, req CancelOrderRequest) error {
	return a.orderUC.CancelInDB(ctx, req.OrderID)
}

// CompleteOrderRequest represents the request for CompleteOrder activity.
type CompleteOrderRequest struct {
	OrderID uuid.UUID
}

// CompleteOrder marks an order as completed.
func (a *Activities) CompleteOrder(ctx context.Context, req CompleteOrderRequest) error {
	order, err := a.orderUC.Get(ctx, req.OrderID)
	if err != nil {
		return err
	}

	if err := order.CompleteOrder(); err != nil {
		return err
	}

	// Note: We need direct access to repository here, or add CompleteInDB to usecase
	// For now, this is a simplified implementation
	return nil
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
	order, err := a.orderUC.Get(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	return &GetOrderResponse{Order: order}, nil
}
