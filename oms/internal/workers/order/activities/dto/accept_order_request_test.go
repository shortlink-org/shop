package dto_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/workers/order/activities/dto"
)

func TestAcceptOrderRequestFromOrder(t *testing.T) {
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	customerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	pickupAddr, err := address.NewAddress("123 Warehouse St", "Moscow", "101000", "Russia")
	require.NoError(t, err)
	deliveryAddr, err := address.NewAddress("456 Customer St", "Moscow", "102000", "Russia")
	require.NoError(t, err)

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	period := orderv1.NewDeliveryPeriod(startTime, endTime)
	pkgInfo := orderv1.NewPackageInfo(2.5)
	recipient := orderv1.NewRecipientContacts("Jane Doe", "+79001234567", "jane@example.com")
	deliveryInfo := orderv1.NewDeliveryInfo(
		pickupAddr, deliveryAddr, period, pkgInfo,
		orderv1.DeliveryPriorityUrgent, &recipient,
	)

	order := orderv1.NewOrderStateFromPersisted(
		orderID, customerID, nil, orderv1.OrderStatus_ORDER_STATUS_PROCESSING, 0, &deliveryInfo,
	)

	req, err := dto.AcceptOrderRequestFromOrder(order)
	require.NoError(t, err)

	require.Equal(t, orderID, req.OrderID)
	require.Equal(t, customerID, req.CustomerID)
	require.Equal(t, "123 Warehouse St", req.PickupAddress.Street)
	require.Equal(t, "Moscow", req.PickupAddress.City)
	require.Equal(t, "101000", req.PickupAddress.PostalCode)
	require.Equal(t, "Russia", req.PickupAddress.Country)
	require.Equal(t, "456 Customer St", req.DeliveryAddress.Street)
	require.Equal(t, "Moscow", req.DeliveryAddress.City)
	require.Equal(t, "102000", req.DeliveryAddress.PostalCode)
	require.Equal(t, "Russia", req.DeliveryAddress.Country)
	require.True(t, req.DeliveryPeriod.StartTime.Equal(startTime))
	require.True(t, req.DeliveryPeriod.EndTime.Equal(endTime))
	require.Equal(t, 2.5, req.PackageInfo.WeightKg)
	require.Equal(t, ports.DeliveryPriorityUrgent, req.Priority)
	require.Equal(t, "Jane Doe", req.RecipientName)
	require.Equal(t, "+79001234567", req.RecipientPhone)
	require.Equal(t, "jane@example.com", req.RecipientEmail)
}

func TestAcceptOrderRequestFromOrder_NoDeliveryInfo(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	order := orderv1.NewOrderStateFromPersisted(
		orderID, customerID, nil, orderv1.OrderStatus_ORDER_STATUS_PENDING, 0, nil,
	)

	_, err := dto.AcceptOrderRequestFromOrder(order)
	require.Error(t, err)
	require.ErrorIs(t, err, dto.ErrNoDeliveryInfo)
}
