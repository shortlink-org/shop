package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	orderCancel "github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	orderGet "github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"
	"github.com/shortlink-org/shop/oms/internal/workers/order/activities/dto"
)

// cancelHandler handles CancelOrder commands (allows mocks in tests).
type cancelHandler interface {
	Handle(ctx context.Context, cmd orderCancel.Command) error
}

// getHandler handles GetOrder queries (allows mocks in tests).
type getHandler interface {
	Handle(ctx context.Context, q orderGet.Query) (orderGet.Result, error)
}

// Activities wraps order command/query handlers for Temporal activities.
// Activities are the bridge between Temporal workflows and application use cases.
// Temporal workflows must never access repositories directly - only through activities.
//
// Note: In the event-driven architecture, order creation happens before the workflow starts
// (CreateOrder command handler publishes event, which triggers the workflow).
// Activities are used for compensation (cancel) and queries during workflow execution.
type Activities struct {
	cancelHandler  cancelHandler
	getHandler     getHandler
	deliveryClient ports.DeliveryClient
}

// New creates a new Activities instance.
func New(
	cancelHandler cancelHandler,
	getHandler getHandler,
	deliveryClient ports.DeliveryClient,
) *Activities {
	return &Activities{
		cancelHandler:  cancelHandler,
		getHandler:     getHandler,
		deliveryClient: deliveryClient,
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

// RequestDeliveryRequest represents the request for RequestDelivery activity.
// The activity loads the order from the repository and uses the domain aggregate's delivery info.
// OrderID and CustomerID for the delivery port are taken from the loaded order (single source of truth).
type RequestDeliveryRequest struct {
	OrderID uuid.UUID
}

// RequestDeliveryResponse represents the response from RequestDelivery activity.
type RequestDeliveryResponse struct {
	PackageID string
	Status    string
}

// ErrDeliveryClientNotConfigured is returned when the delivery client is nil.
var ErrDeliveryClientNotConfigured = errors.New("delivery client not configured")

// ErrOrderHasNoDeliveryInfo is returned when the order has no delivery info set.
var ErrOrderHasNoDeliveryInfo = errors.New("order has no delivery info")

// RequestDelivery sends the order to the Delivery service for processing.
// It loads the order (domain aggregate), maps delivery info to AcceptOrderRequest, and calls the Delivery client.
func (a *Activities) RequestDelivery(ctx context.Context, req RequestDeliveryRequest) (*RequestDeliveryResponse, error) {
	if a.deliveryClient == nil {
		return nil, ErrDeliveryClientNotConfigured
	}

	order, err := a.getHandler.Handle(ctx, orderGet.NewQuery(req.OrderID))
	if err != nil {
		return nil, fmt.Errorf("failed to load order: %w", err)
	}

	if !order.HasDeliveryInfo() {
		return nil, ErrOrderHasNoDeliveryInfo
	}

	deliveryReq, err := dto.AcceptOrderRequestFromOrder(order)
	if err != nil {
		return nil, fmt.Errorf("map order to delivery request: %w", err)
	}

	resp, err := a.deliveryClient.AcceptOrder(ctx, deliveryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to request delivery: %w", err)
	}

	return &RequestDeliveryResponse{
		PackageID: resp.PackageID,
		Status:    resp.Status,
	}, nil
}
