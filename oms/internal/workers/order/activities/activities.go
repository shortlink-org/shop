package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	orderCancel "github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	orderRequestDelivery "github.com/shortlink-org/shop/oms/internal/usecases/order/command/request_delivery"
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

// requestDeliveryHandler records a successful delivery request inside OMS.
type requestDeliveryHandler interface {
	Handle(ctx context.Context, cmd orderRequestDelivery.Command) error
}

// Activities wraps order command/query handlers for Temporal activities.
// Activities are the bridge between Temporal workflows and application use cases.
// Temporal workflows must never access repositories directly - only through activities.
//
// Note: In the event-driven architecture, order creation happens before the workflow starts
// (CreateOrder command handler publishes event, which triggers the workflow).
// Activities are used for compensation (cancel) and queries during workflow execution.
type Activities struct {
	cancelHandler          cancelHandler
	getHandler             getHandler
	requestDeliveryHandler requestDeliveryHandler
	deliveryClient         ports.DeliveryClient
}

const (
	requestDeliveryValidationErrorType = "OrderRequestDeliveryValidationError"
	requestDeliveryConfigErrorType     = "OrderRequestDeliveryConfigError"
	requestDeliveryContractErrorType   = "OrderRequestDeliveryContractError"
	requestDeliveryHeartbeatInterval   = 2 * time.Second
)

// New creates a new Activities instance.
func New(
	cancelHandler cancelHandler,
	getHandler getHandler,
	requestDeliveryHandler requestDeliveryHandler,
	deliveryClient ports.DeliveryClient,
) *Activities {
	return &Activities{
		cancelHandler:          cancelHandler,
		getHandler:             getHandler,
		requestDeliveryHandler: requestDeliveryHandler,
		deliveryClient:         deliveryClient,
	}
}

// NewWithHandlers is a DI-friendly constructor that accepts concrete order handlers.
func NewWithHandlers(
	cancelHandler *orderCancel.Handler,
	getHandler *orderGet.Handler,
	requestDeliveryHandler *orderRequestDelivery.Handler,
	deliveryClient ports.DeliveryClient,
) *Activities {
	return New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
}

// CancelOrderRequest represents the request for CancelOrder activity.
type CancelOrderRequest struct {
	OrderID uuid.UUID
}

// CancelOrder cancels an order in the database.
// This is used for compensation in saga patterns.
func (a *Activities) CancelOrder(ctx context.Context, req CancelOrderRequest) error {
	cmd := orderCancel.NewCommand(req.OrderID)
	err := a.cancelHandler.Handle(ctx, cmd)
	if err == nil {
		return nil
	}

	if isOrderValidationError(err) {
		return temporal.NewNonRetryableApplicationError(err.Error(), requestDeliveryValidationErrorType, err)
	}

	return err
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

// ErrInvalidPackageID is returned when delivery service returns a malformed package ID.
var ErrInvalidPackageID = errors.New("delivery service returned invalid package_id")

// RequestDelivery sends the order to the Delivery service for processing.
// It loads the order (domain aggregate), maps delivery info to AcceptOrderRequest, and calls the Delivery client.
func (a *Activities) RequestDelivery(ctx context.Context, req RequestDeliveryRequest) (*RequestDeliveryResponse, error) {
	if a.deliveryClient == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			ErrDeliveryClientNotConfigured.Error(),
			requestDeliveryConfigErrorType,
			ErrDeliveryClientNotConfigured,
		)
	}

	order, err := a.getHandler.Handle(ctx, orderGet.NewQuery(req.OrderID))
	if err != nil {
		return nil, fmt.Errorf("failed to load order: %w", err)
	}

	if !order.HasDeliveryInfo() {
		return nil, temporal.NewNonRetryableApplicationError(
			ErrOrderHasNoDeliveryInfo.Error(),
			requestDeliveryValidationErrorType,
			ErrOrderHasNoDeliveryInfo,
		)
	}

	deliveryReq, err := dto.AcceptOrderRequestFromOrder(order)
	if err != nil {
		wrappedErr := fmt.Errorf("map order to delivery request: %w", err)
		if isOrderValidationError(err) {
			return nil, temporal.NewNonRetryableApplicationError(
				wrappedErr.Error(),
				requestDeliveryValidationErrorType,
				wrappedErr,
			)
		}

		return nil, wrappedErr
	}

	resp, err := a.acceptOrderWithHeartbeat(ctx, deliveryReq)
	if err != nil {
		if nonRetryableErr := classifyDeliveryAcceptOrderError(err); nonRetryableErr != nil {
			return nil, nonRetryableErr
		}

		return nil, fmt.Errorf("failed to request delivery: %w", err)
	}

	packageID, err := uuid.Parse(resp.PackageID)
	if err != nil {
		wrappedErr := fmt.Errorf("%w: %s", ErrInvalidPackageID, resp.PackageID)

		return nil, temporal.NewNonRetryableApplicationError(
			wrappedErr.Error(),
			requestDeliveryContractErrorType,
			wrappedErr,
		)
	}

	if err := a.requestDeliveryHandler.Handle(
		ctx,
		orderRequestDelivery.NewCommand(req.OrderID, packageID, time.Now().UTC()),
	); err != nil {
		wrappedErr := fmt.Errorf("failed to persist delivery request: %w", err)
		if isOrderValidationError(err) {
			return nil, temporal.NewNonRetryableApplicationError(
				wrappedErr.Error(),
				requestDeliveryValidationErrorType,
				wrappedErr,
			)
		}

		return nil, wrappedErr
	}

	return &RequestDeliveryResponse{
		PackageID: resp.PackageID,
		Status:    resp.Status,
	}, nil
}

type acceptOrderResult struct {
	response *ports.AcceptOrderResponse
	err      error
}

func (a *Activities) acceptOrderWithHeartbeat(
	ctx context.Context,
	req ports.AcceptOrderRequest,
) (*ports.AcceptOrderResponse, error) {
	safeRecordHeartbeat(ctx, "request-delivery:start")

	resultCh := make(chan acceptOrderResult, 1)
	go func() {
		resp, err := a.deliveryClient.AcceptOrder(ctx, req)
		resultCh <- acceptOrderResult{response: resp, err: err}
	}()

	ticker := time.NewTicker(requestDeliveryHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultCh:
			safeRecordHeartbeat(ctx, "request-delivery:finish")
			return result.response, result.err
		case <-ticker.C:
			safeRecordHeartbeat(ctx, "request-delivery:waiting")
		}
	}
}

func safeRecordHeartbeat(ctx context.Context, details ...interface{}) {
	defer func() {
		if recovered := recover(); recovered != nil && fmt.Sprint(recovered) != "getActivityOutboundInterceptor: Not an activity context" {
			panic(recovered)
		}
	}()

	activity.RecordHeartbeat(ctx, details...)
}

func classifyDeliveryAcceptOrderError(err error) error {
	wrappedErr := fmt.Errorf("failed to request delivery: %w", err)

	switch grpcstatus.Code(err) {
	case codes.InvalidArgument, codes.FailedPrecondition:
		return temporal.NewNonRetryableApplicationError(
			wrappedErr.Error(),
			requestDeliveryValidationErrorType,
			wrappedErr,
		)
	case codes.AlreadyExists:
		return temporal.NewNonRetryableApplicationError(
			wrappedErr.Error(),
			requestDeliveryContractErrorType,
			wrappedErr,
		)
	case codes.Unauthenticated, codes.PermissionDenied, codes.Unimplemented:
		return temporal.NewNonRetryableApplicationError(
			wrappedErr.Error(),
			requestDeliveryConfigErrorType,
			wrappedErr,
		)
	default:
		return nil
	}
}

func isOrderValidationError(err error) bool {
	var (
		orderTerminalStateErr        *orderv1.OrderTerminalStateError
		deliveryAlreadyRequestedErr  *orderv1.DeliveryAlreadyRequestedError
		deliveryAlreadyInProgressErr *orderv1.DeliveryAlreadyInProgressError
		invalidOrderTransitionErr    *orderv1.InvalidOrderTransitionError
		invalidDeliveryTransitionErr *orderv1.InvalidDeliveryStatusTransitionError
	)

	return errors.Is(err, ErrOrderHasNoDeliveryInfo) ||
		errors.Is(err, ErrInvalidPackageID) ||
		errors.Is(err, dto.ErrNoDeliveryInfo) ||
		errors.Is(err, dto.ErrUnsupportedDeliveryPriority) ||
		errors.Is(err, orderv1.ErrInvalidDeliveryInfo) ||
		errors.Is(err, orderv1.ErrDeliveryInfoRequired) ||
		errors.Is(err, orderv1.ErrOrderInvalidStateTransition) ||
		errors.As(err, &orderTerminalStateErr) ||
		errors.As(err, &deliveryAlreadyRequestedErr) ||
		errors.As(err, &deliveryAlreadyInProgressErr) ||
		errors.As(err, &invalidOrderTransitionErr) ||
		errors.As(err, &invalidDeliveryTransitionErr)
}
