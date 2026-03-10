package activities

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	orderCancel "github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	orderRequestDelivery "github.com/shortlink-org/shop/oms/internal/usecases/order/command/request_delivery"
	orderGet "github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"
)

// mockRequestDeliveryHandler is a mock for RequestDelivery command (used in RequestDelivery activity).
type mockRequestDeliveryHandler struct {
	mock.Mock
}

func (m *mockRequestDeliveryHandler) Handle(ctx context.Context, cmd orderRequestDelivery.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

// mockDeliveryClient is a mock for ports.DeliveryClient (used in RequestDelivery activity).
type mockDeliveryClient struct {
	mock.Mock
}

func (m *mockDeliveryClient) AcceptOrder(ctx context.Context, req ports.AcceptOrderRequest) (*ports.AcceptOrderResponse, error) {
	args := m.Called(ctx, req)

	var err error
	switch value := args.Get(1).(type) {
	case nil:
	case func(context.Context, ports.AcceptOrderRequest) error:
		err = value(ctx, req)
	case error:
		err = value
	default:
		panic("unexpected error return type from mockDeliveryClient.AcceptOrder")
	}

	if args.Get(0) == nil {
		return nil, err
	}

	switch value := args.Get(0).(type) {
	case func(context.Context, ports.AcceptOrderRequest) *ports.AcceptOrderResponse:
		return value(ctx, req), err
	case *ports.AcceptOrderResponse:
		return value, err
	default:
		return nil, err
	}
}

// mockCancelHandler is a mock implementation of CommandHandler for cancel command.
type mockCancelHandler struct {
	mock.Mock
}

func (m *mockCancelHandler) Handle(ctx context.Context, cmd orderCancel.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

// mockGetHandler is a mock implementation of QueryHandler for get query.
type mockGetHandler struct {
	mock.Mock
}

func (m *mockGetHandler) Handle(ctx context.Context, query orderGet.Query) (orderGet.Result, error) {
	args := m.Called(ctx, query)
	err := args.Error(1)

	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*orderv1.OrderState)
	if !ok {
		return nil, err
	}

	return res, err
}

// Fixed UUIDs for consistent testing
var (
	testOrderID    = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	testCustomerID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
)

func createOrderWithDeliveryInfo(t *testing.T) *orderv1.OrderState {
	t.Helper()

	pickupAddr, err := address.NewAddress("123 Warehouse St", "Moscow", "101000", "Russia")
	require.NoError(t, err)

	deliveryAddr, err := address.NewAddress("456 Customer St", "Moscow", "102000", "Russia")
	require.NoError(t, err)

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	deliveryInfo := orderv1.NewDeliveryInfo(
		pickupAddr,
		deliveryAddr,
		orderv1.NewDeliveryPeriod(startTime, endTime),
		orderv1.NewPackageInfo(2.5),
		orderv1.DeliveryPriorityNormal,
		nil,
	)

	return orderv1.NewOrderStateFromPersisted(
		testOrderID,
		testCustomerID,
		nil,
		orderv1.OrderStatus_ORDER_STATUS_PROCESSING,
		0,
		&deliveryInfo,
		commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
		nil,
	)
}

func TestActivities_CancelOrder_Success(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)

	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)

	// Set up expectation
	cancelHandler.On("Handle", mock.Anything, orderCancel.NewCommand(testOrderID)).Return(nil)

	// Execute activity
	err := activities.CancelOrder(context.Background(), CancelOrderRequest{
		OrderID: testOrderID,
	})

	// Assert
	require.NoError(t, err)
	cancelHandler.AssertExpectations(t)
}

func TestActivities_CancelOrder_HandlerError(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)

	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)

	// Set up expectation with error
	expectedErr := errors.New("order not found")
	cancelHandler.On("Handle", mock.Anything, orderCancel.NewCommand(testOrderID)).Return(expectedErr)

	// Execute activity
	err := activities.CancelOrder(context.Background(), CancelOrderRequest{
		OrderID: testOrderID,
	})

	// Assert
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
	cancelHandler.AssertExpectations(t)
}

func TestActivities_GetOrder_Success(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)

	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)

	// Create expected order state
	expectedOrder := orderv1.NewOrderState(testCustomerID)

	// Set up expectation
	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(expectedOrder, nil)

	// Execute activity
	response, err := activities.GetOrder(context.Background(), GetOrderRequest{
		OrderID: testOrderID,
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Order)
	require.Equal(t, testCustomerID, response.Order.GetCustomerId())
	getHandler.AssertExpectations(t)
}

func TestActivities_GetOrder_NotFound(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)

	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)

	// Set up expectation with error
	expectedErr := errors.New("order not found")
	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(nil, expectedErr)

	// Execute activity
	response, err := activities.GetOrder(context.Background(), GetOrderRequest{
		OrderID: testOrderID,
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, response)
	require.Equal(t, expectedErr, err)
	getHandler.AssertExpectations(t)
}

func TestActivities_GetOrder_ContextCancelled(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)

	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)

	// Create canceled context
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errors.New("test context canceled"))

	// Set up expectation with context error
	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(nil, context.Canceled)

	// Execute activity
	response, err := activities.GetOrder(ctx, GetOrderRequest{
		OrderID: testOrderID,
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorIs(t, err, context.Canceled)
	getHandler.AssertExpectations(t)
}

func TestActivities_New(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)

	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)

	require.NotNil(t, activities)
}

func TestActivities_RequestDelivery_DeliveryClientNotConfigured(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, nil)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrDeliveryClientNotConfigured)
	var appErr *temporal.ApplicationError
	require.ErrorAs(t, err, &appErr)
	require.True(t, appErr.NonRetryable())
	require.Equal(t, requestDeliveryConfigErrorType, appErr.Type())
}

func TestActivities_RequestDelivery_NoDeliveryInfo(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := orderv1.NewOrderState(testCustomerID)
	order.SetID(testOrderID)

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrOrderHasNoDeliveryInfo)
	var appErr *temporal.ApplicationError
	require.ErrorAs(t, err, &appErr)
	require.True(t, appErr.NonRetryable())
	require.Equal(t, requestDeliveryValidationErrorType, appErr.Type())
	getHandler.AssertExpectations(t)
	deliveryClient.AssertNotCalled(t, "AcceptOrder", mock.Anything, mock.Anything)
	requestDeliveryHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func TestActivities_RequestDelivery_AcceptOrderError(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)
	expectedErr := errors.New("delivery backend unavailable")

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(nil, expectedErr)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorContains(t, err, "failed to request delivery")
	require.ErrorContains(t, err, expectedErr.Error())
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func TestActivities_RequestDelivery_AcceptOrderInvalidArgumentIsNonRetryable(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(
		nil,
		grpcstatus.Error(codes.InvalidArgument, "invalid delivery request"),
	)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	var appErr *temporal.ApplicationError
	require.ErrorAs(t, err, &appErr)
	require.True(t, appErr.NonRetryable())
	require.Equal(t, requestDeliveryValidationErrorType, appErr.Type())
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func TestActivities_RequestDelivery_AcceptOrderAlreadyExistsIsNonRetryable(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(
		nil,
		grpcstatus.Error(codes.AlreadyExists, "package for order already exists"),
	)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	var appErr *temporal.ApplicationError
	require.ErrorAs(t, err, &appErr)
	require.True(t, appErr.NonRetryable())
	require.Equal(t, requestDeliveryContractErrorType, appErr.Type())
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func TestActivities_RequestDelivery_InvalidPackageID(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(&ports.AcceptOrderResponse{
		PackageID: "not-a-uuid",
		Status:    "ACCEPTED",
	}, nil)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrInvalidPackageID)
	var appErr *temporal.ApplicationError
	require.ErrorAs(t, err, &appErr)
	require.True(t, appErr.NonRetryable())
	require.Equal(t, requestDeliveryContractErrorType, appErr.Type())
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func TestActivities_RequestDelivery_ContextCancelled(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)

	ctx, cancel := context.WithCancel(context.Background())
	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(
		func(callCtx context.Context, _ ports.AcceptOrderRequest) *ports.AcceptOrderResponse {
			<-callCtx.Done()
			return nil
		},
		func(callCtx context.Context, _ ports.AcceptOrderRequest) error {
			return callCtx.Err()
		},
	)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	response, err := activities.RequestDelivery(ctx, RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorIs(t, err, context.Canceled)
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
}

func TestActivities_RequestDelivery_PersistError(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)
	packageID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174777")
	expectedErr := errors.New("cannot persist request")

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(&ports.AcceptOrderResponse{
		PackageID: packageID.String(),
		Status:    "ACCEPTED",
	}, nil)
	requestDeliveryHandler.On("Handle", mock.Anything, mock.MatchedBy(func(cmd orderRequestDelivery.Command) bool {
		return cmd.OrderID == testOrderID && cmd.PackageID == packageID && !cmd.RequestedAt.IsZero()
	})).Return(expectedErr)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.Error(t, err)
	require.Nil(t, response)
	require.ErrorContains(t, err, "failed to persist delivery request")
	require.ErrorContains(t, err, expectedErr.Error())
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertExpectations(t)
}

func TestActivities_RequestDelivery_Success(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)
	deliveryClient := new(mockDeliveryClient)
	requestDeliveryHandler := new(mockRequestDeliveryHandler)
	activities := New(cancelHandler, getHandler, requestDeliveryHandler, deliveryClient)
	order := createOrderWithDeliveryInfo(t)
	packageID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174888")

	getHandler.On("Handle", mock.Anything, orderGet.NewQuery(testOrderID)).Return(order, nil)
	deliveryClient.On("AcceptOrder", mock.Anything, mock.Anything).Return(&ports.AcceptOrderResponse{
		PackageID: packageID.String(),
		Status:    "ACCEPTED",
	}, nil)
	requestDeliveryHandler.On("Handle", mock.Anything, mock.MatchedBy(func(cmd orderRequestDelivery.Command) bool {
		return cmd.OrderID == testOrderID && cmd.PackageID == packageID && !cmd.RequestedAt.IsZero()
	})).Return(nil)

	response, err := activities.RequestDelivery(context.Background(), RequestDeliveryRequest{
		OrderID: testOrderID,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, packageID.String(), response.PackageID)
	require.Equal(t, "ACCEPTED", response.Status)
	getHandler.AssertExpectations(t)
	deliveryClient.AssertExpectations(t)
	requestDeliveryHandler.AssertExpectations(t)
}
