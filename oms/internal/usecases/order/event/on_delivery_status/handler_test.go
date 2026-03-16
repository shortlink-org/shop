package on_delivery_status

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
)

func TestIsDuplicateOrStale(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		currentStatus  commonv1.DeliveryStatus
		eventType      kafka.DeliveryEventType
		expectedResult bool
	}{
		{
			name:           "duplicate target is ignored",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
			eventType:      kafka.EventTypePackageAssigned,
			expectedResult: true,
		},
		{
			name:           "assigned is allowed when accepted was skipped",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
			eventType:      kafka.EventTypePackageAssigned,
			expectedResult: false,
		},
		{
			name:           "in transit is allowed when earlier statuses were skipped",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
			eventType:      kafka.EventTypePackageInTransit,
			expectedResult: false,
		},
		{
			name:           "assigned after in transit is stale",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
			eventType:      kafka.EventTypePackageAssigned,
			expectedResult: true,
		},
		{
			name:           "accepted after assigned is stale",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
			eventType:      kafka.EventTypePackageAccepted,
			expectedResult: true,
		},
		{
			name:           "delivered is allowed directly from assigned",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
			eventType:      kafka.EventTypePackageDelivered,
			expectedResult: false,
		},
		{
			name:           "terminal cross-over is stale",
			currentStatus:  commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED,
			eventType:      kafka.EventTypePackageNotDelivered,
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			order := newOrderStateForDeliveryStatus(t, tc.currentStatus)

			require.Equal(t, tc.expectedResult, isDuplicateOrStale(order, tc.eventType))
		})
	}
}

func newOrderStateForDeliveryStatus(t *testing.T, status commonv1.DeliveryStatus) *orderv1.OrderState {
	t.Helper()

	return orderv1.NewOrderStateFromPersisted(
		uuid.New(),
		uuid.New(),
		orderv1.Items{
			orderv1.NewItem(uuid.New(), 1, decimal.NewFromInt(10)),
		},
		orderv1.OrderStatus_ORDER_STATUS_PROCESSING,
		1,
		nil,
		status,
		nil,
	)
}
