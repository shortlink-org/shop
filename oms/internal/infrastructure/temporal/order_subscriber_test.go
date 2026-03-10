package temporal

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	sdklogger "github.com/shortlink-org/go-sdk/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	temporalmocks "go.temporal.io/sdk/mocks"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/events/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	queuev1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/events"
	orderrepomocks "github.com/shortlink-org/shop/oms/internal/usecases/order/command/create_order_from_cart/mocks"
)

func TestOrderEventSubscriber_Register_StartsWorkflowOnOrderCreated(t *testing.T) {
	t.Parallel()

	log := newDiscardLogger(t)
	temporalClient := new(temporalmocks.Client)
	orderRepo := new(orderrepomocks.MockOrderRepository)
	publisher := events.NewInMemoryPublisher()
	subscriber := NewOrderEventSubscriber(log, temporalClient, orderRepo)
	subscriber.Register(publisher)

	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	order := createOrderStateForSubscriber(t, orderID, customerID, true)

	orderRepo.EXPECT().Load(mock.Anything, orderID).Return(order, nil).Once()
	temporalClient.On(
		"ExecuteWorkflow",
		mock.Anything,
		mock.MatchedBy(func(opts client.StartWorkflowOptions) bool {
			return opts.ID == "order-"+orderID.String() &&
				opts.TaskQueue == GetQueueName(queuev1.OrderTaskQueue) &&
				opts.WorkflowExecutionTimeout == 24*time.Hour
		}),
		OrderWorkflowName,
		orderID,
		customerID,
		mock.Anything,
		true,
	).Run(func(args mock.Arguments) {
		itemsArg, ok := args.Get(5).(orderv1.Items)
		require.True(t, ok)
		require.Len(t, itemsArg, 1)
		require.Equal(t, order.GetItems()[0].GetGoodId(), itemsArg[0].GetGoodId())
		require.Equal(t, order.GetItems()[0].GetQuantity(), itemsArg[0].GetQuantity())
		require.True(t, order.GetItems()[0].GetPrice().Equal(itemsArg[0].GetPrice()))
	}).Return(nil, nil).Once()

	err := publisher.Publish(context.Background(), &eventsv1.OrderCreated{
		OrderId:    orderID.String(),
		CustomerId: customerID.String(),
	})
	require.NoError(t, err)

	orderRepo.AssertExpectations(t)
	temporalClient.AssertExpectations(t)
}

func TestOrderEventSubscriber_Register_SignalsWorkflowOnOrderCancelled(t *testing.T) {
	t.Parallel()

	log := newDiscardLogger(t)
	temporalClient := new(temporalmocks.Client)
	orderRepo := new(orderrepomocks.MockOrderRepository)
	publisher := events.NewInMemoryPublisher()
	subscriber := NewOrderEventSubscriber(log, temporalClient, orderRepo)
	subscriber.Register(publisher)

	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	temporalClient.On(
		"SignalWorkflow",
		mock.Anything,
		"order-"+orderID.String(),
		"",
		"cancel",
		"DELIVERY_FAILED",
	).Return(nil).Once()

	err := publisher.Publish(context.Background(), &eventsv1.OrderCancelled{
		OrderId: orderID.String(),
		Reason:  "DELIVERY_FAILED",
	})
	require.NoError(t, err)

	temporalClient.AssertExpectations(t)
}

func createOrderStateForSubscriber(
	t *testing.T,
	orderID uuid.UUID,
	customerID uuid.UUID,
	withDelivery bool,
) *orderv1.OrderState {
	t.Helper()

	items := orderv1.Items{
		orderv1.NewItem(
			uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
			2,
			decimal.RequireFromString("19.99"),
		),
	}

	var deliveryInfo *orderv1.DeliveryInfo
	if withDelivery {
		pickupAddr, err := address.NewAddress("123 Warehouse St", "Moscow", "101000", "Russia")
		require.NoError(t, err)
		deliveryAddr, err := address.NewAddress("456 Customer St", "Moscow", "102000", "Russia")
		require.NoError(t, err)

		info := orderv1.NewDeliveryInfo(
			pickupAddr,
			deliveryAddr,
			orderv1.NewDeliveryPeriod(time.Now().Add(24*time.Hour), time.Now().Add(26*time.Hour)),
			orderv1.NewPackageInfo(2.5),
			orderv1.DeliveryPriorityNormal,
			nil,
		)
		deliveryInfo = &info
	}

	return orderv1.NewOrderStateFromPersisted(
		orderID,
		customerID,
		items,
		orderv1.OrderStatus_ORDER_STATUS_PROCESSING,
		1,
		deliveryInfo,
		commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
		nil,
	)
}

func newDiscardLogger(t *testing.T) sdklogger.Logger {
	t.Helper()

	cfg := sdklogger.Default()
	cfg.Writer = io.Discard

	log, err := sdklogger.New(cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, log.Close())
	})

	return log
}
