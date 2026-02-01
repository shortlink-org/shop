package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	orderCancel "github.com/shortlink-org/shop/oms/internal/usecases/order/command/cancel"
	orderGet "github.com/shortlink-org/shop/oms/internal/usecases/order/query/get"
)

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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orderv1.OrderState), args.Error(1)
}

// Fixed UUIDs for consistent testing
var (
	testOrderID    = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	testCustomerID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
)

func TestActivities_CancelOrder_Success(t *testing.T) {
	cancelHandler := new(mockCancelHandler)
	getHandler := new(mockGetHandler)

	activities := New(cancelHandler, getHandler)

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

	activities := New(cancelHandler, getHandler)

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

	activities := New(cancelHandler, getHandler)

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

	activities := New(cancelHandler, getHandler)

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

	activities := New(cancelHandler, getHandler)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

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

	activities := New(cancelHandler, getHandler)

	require.NotNil(t, activities)
}
