package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
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
	cancelHandler  ports.CommandHandler[orderCancel.Command]
	getHandler     ports.QueryHandler[orderGet.Query, orderGet.Result]
	deliveryClient ports.DeliveryClient
}

// New creates a new Activities instance.
func New(
	cancelHandler ports.CommandHandler[orderCancel.Command],
	getHandler ports.QueryHandler[orderGet.Query, orderGet.Result],
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
type RequestDeliveryRequest struct {
	OrderID         uuid.UUID
	CustomerID      uuid.UUID
	PickupAddress   DeliveryAddressDTO
	DeliveryAddress DeliveryAddressDTO
	StartTime       time.Time
	EndTime         time.Time
	WeightKg        float64
	Dimensions      string
	Priority        int32 // 0=unspecified, 1=normal, 2=urgent
}

// DeliveryAddressDTO represents address data for delivery request.
type DeliveryAddressDTO struct {
	Street     string
	City       string
	PostalCode string
	Country    string
	Latitude   float64
	Longitude  float64
}

// RequestDeliveryResponse represents the response from RequestDelivery activity.
type RequestDeliveryResponse struct {
	PackageID string
	Status    string
}

// RequestDelivery sends the order to the Delivery service for processing.
// This activity calls the Delivery service's AcceptOrder gRPC endpoint.
func (a *Activities) RequestDelivery(ctx context.Context, req RequestDeliveryRequest) (*RequestDeliveryResponse, error) {
	if a.deliveryClient == nil {
		return nil, fmt.Errorf("delivery client not configured")
	}

	deliveryReq := ports.AcceptOrderRequest{
		OrderID:    req.OrderID,
		CustomerID: req.CustomerID,
		PickupAddress: ports.DeliveryAddress{
			Street:     req.PickupAddress.Street,
			City:       req.PickupAddress.City,
			PostalCode: req.PickupAddress.PostalCode,
			Country:    req.PickupAddress.Country,
			Latitude:   req.PickupAddress.Latitude,
			Longitude:  req.PickupAddress.Longitude,
		},
		DeliveryAddress: ports.DeliveryAddress{
			Street:     req.DeliveryAddress.Street,
			City:       req.DeliveryAddress.City,
			PostalCode: req.DeliveryAddress.PostalCode,
			Country:    req.DeliveryAddress.Country,
			Latitude:   req.DeliveryAddress.Latitude,
			Longitude:  req.DeliveryAddress.Longitude,
		},
		DeliveryPeriod: ports.DeliveryPeriodDTO{
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
		},
		PackageInfo: ports.PackageInfoDTO{
			WeightKg:   req.WeightKg,
			Dimensions: req.Dimensions,
		},
		Priority: ports.DeliveryPriorityDTO(req.Priority),
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
