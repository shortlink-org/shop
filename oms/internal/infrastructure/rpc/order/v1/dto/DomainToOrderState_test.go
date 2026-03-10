package dto

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
)

func TestDomainToOrderState_MapsDeliveryLifecycleFields(t *testing.T) {
	t.Parallel()

	customerID := uuid.New()
	goodID := uuid.New()
	packageID := uuid.New()
	requestedAt := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)

	order := orderv1.NewOrderState(customerID)
	require.NoError(t, order.SetDeliveryInfo(testDeliveryInfo(t)))
	require.NoError(t, order.CreateOrder(context.Background(), orderv1.Items{
		orderv1.NewItem(goodID, 2, decimal.NewFromFloat(19.99)),
	}))
	order.ClearDomainEvents()
	require.NoError(t, order.RequestDelivery(&packageID, requestedAt))
	require.NoError(t, order.ApplyDeliveryAccepted(&packageID, requestedAt.Add(time.Minute)))

	out := DomainToOrderState(order)
	require.NotNil(t, out)
	require.Equal(t, customerID.String(), out.GetCustomerId())
	require.Equal(t, commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED, out.GetDeliveryStatus())
	require.Equal(t, packageID.String(), out.GetPackageId())
	require.NotNil(t, out.GetRequestedAt())
	require.True(t, out.GetRequestedAt().AsTime().Equal(requestedAt))
	require.NotNil(t, out.GetDeliveryInfo())
	require.Equal(t, "123 Warehouse St", out.GetDeliveryInfo().GetPickupAddress().GetStreet())
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
