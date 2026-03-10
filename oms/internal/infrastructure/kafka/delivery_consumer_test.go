package kafka

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	deliverycommon "github.com/shortlink-org/shop/oms/internal/domain/delivery/common/v1"
	deliveryevents "github.com/shortlink-org/shop/oms/internal/domain/delivery/events/v1"
)

func TestDeliveryConsumer_UnmarshalDeliveryEvent_PackageAssigned(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.March, 11, 9, 30, 0, 0, time.UTC)
	event := &deliveryevents.PackageAssignedEvent{
		PackageId:  uuid.NewString(),
		CourierId:  uuid.NewString(),
		Status:     deliverycommon.PackageStatus_PACKAGE_STATUS_ASSIGNED,
		AssignedAt: timestamppb.New(occurredAt.Add(-5 * time.Minute)),
		OccurredAt: timestamppb.New(occurredAt),
	}

	payload, err := proto.Marshal(event)
	require.NoError(t, err)

	statusEvent, err := (&DeliveryConsumer{}).unmarshalDeliveryEvent("PackageAssignedEvent", payload)
	require.NoError(t, err)
	require.Equal(t, event.GetPackageId(), statusEvent.PackageID)
	require.Equal(t, event.GetCourierId(), statusEvent.CourierID)
	require.Equal(t, EventTypePackageAssigned, statusEvent.EventType)
	require.Equal(t, "PACKAGE_STATUS_ASSIGNED", statusEvent.Status)
	require.Equal(t, occurredAt, statusEvent.OccurredAt)
}
