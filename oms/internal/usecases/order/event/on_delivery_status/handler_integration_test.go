//go:build integration

package on_delivery_status

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	cqrsbus "github.com/shortlink-org/go-sdk/cqrs/bus"
	cqrsmessage "github.com/shortlink-org/go-sdk/cqrs/message"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/go-sdk/logger"

	deliverycommon "github.com/shortlink-org/shop/oms/internal/domain/delivery/common/v1"
	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	ordercommon "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	kafkaevent "github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
	orderrepo "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/testhelpers"
	requestdelivery "github.com/shortlink-org/shop/oms/internal/usecases/order/command/request_delivery"
	uowpg "github.com/shortlink-org/shop/oms/pkg/uow/postgres"
)

func TestHandleDeliveryStatus_Integration(t *testing.T) {
	testCases := []struct {
		name                   string
		buildTerminalEvent     func(packageID, courierID uuid.UUID, occurredAt time.Time) kafkaevent.DeliveryStatusEvent
		expectedOrderStatus    orderv1.OrderStatus
		expectedDeliveryStatus ordercommon.DeliveryStatus
	}{
		{
			name: "delivered completes order",
			buildTerminalEvent: func(packageID, courierID uuid.UUID, occurredAt time.Time) kafkaevent.DeliveryStatusEvent {
				return kafkaevent.DeliveryStatusEvent{
					PackageID:  packageID.String(),
					CourierID:  courierID.String(),
					Status:     "PACKAGE_STATUS_DELIVERED",
					EventType:  kafkaevent.EventTypePackageDelivered,
					OccurredAt: occurredAt,
					DeliveryLocation: &deliverycommon.Location{
						Latitude:  55.751244,
						Longitude: 37.618423,
					},
				}
			},
			expectedOrderStatus:    orderv1.OrderStatus_ORDER_STATUS_COMPLETED,
			expectedDeliveryStatus: ordercommon.DeliveryStatus_DELIVERY_STATUS_DELIVERED,
		},
		{
			name: "not delivered cancels order",
			buildTerminalEvent: func(packageID, courierID uuid.UUID, occurredAt time.Time) kafkaevent.DeliveryStatusEvent {
				return kafkaevent.DeliveryStatusEvent{
					PackageID:  packageID.String(),
					CourierID:  courierID.String(),
					Status:     "PACKAGE_STATUS_NOT_DELIVERED",
					EventType:  kafkaevent.EventTypePackageNotDelivered,
					OccurredAt: occurredAt,
					NotDeliveredDetails: &deliverycommon.NotDeliveredDetails{
						Reason:      deliverycommon.NotDeliveredReason_NOT_DELIVERED_REASON_CUSTOMER_NOT_AVAILABLE,
						Description: "customer unavailable",
					},
				}
			},
			expectedOrderStatus:    orderv1.OrderStatus_ORDER_STATUS_CANCELED,
			expectedDeliveryStatus: ordercommon.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			env := setupDeliveryLifecycleTestEnv(t)
			ctx := context.Background()

			orderID, packageID := env.createOrderWithRequestedDelivery(t, ctx)
			courierID := uuid.New()

			require.Equal(t, int64(1), env.outboxCount(t))

			require.NoError(t, env.deliveryHandler.HandleDeliveryStatus(ctx, kafkaevent.DeliveryStatusEvent{
				OrderID:    orderID.String(),
				PackageID:  packageID.String(),
				Status:     "PACKAGE_STATUS_ACCEPTED",
				EventType:  kafkaevent.EventTypePackageAccepted,
				OccurredAt: time.Date(2026, time.March, 11, 10, 1, 0, 0, time.UTC),
			}))
			require.Equal(t, int64(2), env.outboxCount(t))

			require.NoError(t, env.deliveryHandler.HandleDeliveryStatus(ctx, kafkaevent.DeliveryStatusEvent{
				PackageID:  packageID.String(),
				CourierID:  courierID.String(),
				Status:     "PACKAGE_STATUS_ASSIGNED",
				EventType:  kafkaevent.EventTypePackageAssigned,
				OccurredAt: time.Date(2026, time.March, 11, 10, 2, 0, 0, time.UTC),
			}))
			require.Equal(t, int64(3), env.outboxCount(t))

			require.NoError(t, env.deliveryHandler.HandleDeliveryStatus(ctx, kafkaevent.DeliveryStatusEvent{
				PackageID:  packageID.String(),
				CourierID:  courierID.String(),
				Status:     "PACKAGE_STATUS_IN_TRANSIT",
				EventType:  kafkaevent.EventTypePackageInTransit,
				OccurredAt: time.Date(2026, time.March, 11, 10, 3, 0, 0, time.UTC),
			}))
			require.Equal(t, int64(4), env.outboxCount(t))

			require.NoError(t, env.deliveryHandler.HandleDeliveryStatus(
				ctx,
				tc.buildTerminalEvent(packageID, courierID, time.Date(2026, time.March, 11, 10, 4, 0, 0, time.UTC)),
			))
			require.Equal(t, int64(6), env.outboxCount(t))

			order := env.loadOrder(t, orderID)
			require.Equal(t, tc.expectedOrderStatus, order.GetStatus())
			require.Equal(t, tc.expectedDeliveryStatus, order.GetDeliveryStatus())
			require.True(t, order.HasDeliveryRequest())
			require.NotNil(t, order.GetDeliveryInfo())
			require.NotNil(t, order.GetDeliveryInfo().GetPackageId())
			require.Equal(t, packageID, *order.GetDeliveryInfo().GetPackageId())

			loadedByPackage := env.loadOrderByPackageID(t, packageID)
			require.Equal(t, orderID, loadedByPackage.GetOrderID())
			require.Equal(t, tc.expectedOrderStatus, loadedByPackage.GetStatus())
		})
	}
}

type deliveryLifecycleTestEnv struct {
	store           *orderrepo.Store
	uow             *uowpg.UoW
	requestHandler  *requestdelivery.Handler
	deliveryHandler *Handler
	postgres        *testhelpers.PostgresContainer
}

func setupDeliveryLifecycleTestEnv(t *testing.T) *deliveryLifecycleTestEnv {
	t.Helper()

	pc := testhelpers.SetupPostgresContainer(t)
	store, err := orderrepo.New(context.Background(), pc.DB())
	require.NoError(t, err)
	t.Cleanup(store.Close)

	logCfg := logger.Default()
	logCfg.Writer = io.Discard
	logCfg.Level = logger.WARN_LEVEL

	log, err := logger.New(logCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = log.Close() })

	uow := uowpg.New(pc.Pool)
	publisher := newTxAwareEventBus(t)

	requestHandler, err := requestdelivery.NewHandler(log, uow, store, publisher)
	require.NoError(t, err)

	deliveryHandler, err := NewHandler(log, uow, store, publisher)
	require.NoError(t, err)

	return &deliveryLifecycleTestEnv{
		store:           store,
		uow:             uow,
		requestHandler:  requestHandler,
		deliveryHandler: deliveryHandler,
		postgres:        pc,
	}
}

func newTxAwareEventBus(t *testing.T) ports.EventPublisher {
	t.Helper()

	namer := cqrsmessage.NewShortlinkNamer("oms")
	publisherBus, err := cqrsbus.NewEventBusWithOptions(
		nil,
		cqrsmessage.NewJSONMarshaler(namer),
		namer,
		cqrsbus.WithTxAwareOutbox("oms_outbox", watermill.NewStdLogger(false, false)),
	)
	require.NoError(t, err)

	return cqrsbus.NewEventPublisher(publisherBus)
}

func (e *deliveryLifecycleTestEnv) createOrderWithRequestedDelivery(t *testing.T, ctx context.Context) (uuid.UUID, uuid.UUID) {
	t.Helper()

	customerID := uuid.New()
	packageID := uuid.New()

	order := orderv1.NewOrderState(customerID)
	require.NoError(t, order.SetDeliveryInfo(testDeliveryInfo(t)))
	require.NoError(t, order.CreateOrder(ctx, orderv1.Items{
		orderv1.NewItem(uuid.New(), 2, decimal.NewFromFloat(19.99)),
	}))

	txCtx, err := e.uow.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, e.store.Save(txCtx, order))
	require.NoError(t, e.uow.Commit(txCtx))

	requestedAt := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	require.NoError(t, e.requestHandler.Handle(ctx, requestdelivery.NewCommand(order.GetOrderID(), packageID, requestedAt)))

	return order.GetOrderID(), packageID
}

func (e *deliveryLifecycleTestEnv) loadOrder(t *testing.T, orderID uuid.UUID) *orderv1.OrderState {
	t.Helper()

	ctx := context.Background()
	txCtx, err := e.uow.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, e.uow.Rollback(txCtx))
	}()

	order, err := e.store.Load(txCtx, orderID)
	require.NoError(t, err)

	return order
}

func (e *deliveryLifecycleTestEnv) loadOrderByPackageID(t *testing.T, packageID uuid.UUID) *orderv1.OrderState {
	t.Helper()

	ctx := context.Background()
	txCtx, err := e.uow.Begin(ctx)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, e.uow.Rollback(txCtx))
	}()

	order, err := e.store.LoadByPackageID(txCtx, packageID)
	require.NoError(t, err)

	return order
}

func (e *deliveryLifecycleTestEnv) outboxCount(t *testing.T) int64 {
	t.Helper()

	var count int64
	err := e.postgres.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM watermill_oms_outbox`).Scan(&count)
	require.NoError(t, err)

	return count
}

func testDeliveryInfo(t *testing.T) orderv1.DeliveryInfo {
	t.Helper()

	pickupAddr, err := address.NewAddress("123 Warehouse St", "Moscow", "101000", "Russia")
	require.NoError(t, err)

	deliveryAddr, err := address.NewAddress("456 Customer St", "Moscow", "102000", "Russia")
	require.NoError(t, err)

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	return orderv1.NewDeliveryInfo(
		pickupAddr,
		deliveryAddr,
		orderv1.NewDeliveryPeriod(startTime, endTime),
		orderv1.NewPackageInfo(2.5),
		orderv1.DeliveryPriorityNormal,
		nil,
	)
}
