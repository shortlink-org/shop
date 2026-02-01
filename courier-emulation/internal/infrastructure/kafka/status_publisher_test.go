package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

func TestStatusPublisher_PublishPickUp(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	ctx := context.Background()

	event := PickUpOrderEvent{
		PackageID: "pkg-123",
		CourierID: "courier-456",
		PickupLocation: Location{
			Latitude:  52.5200,
			Longitude: 13.4050,
			Accuracy:  10.0,
			Timestamp: time.Now(),
		},
		PickedUpAt: time.Now(),
	}

	err := statusPub.PublishPickUp(ctx, event)
	require.NoError(t, err)

	// Verify message was published
	messages := mockPub.messages[TopicPickUpOrder]
	require.Len(t, messages, 1)

	// Verify message content
	var receivedEvent PickUpOrderEvent
	err = json.Unmarshal(messages[0].Payload, &receivedEvent)
	require.NoError(t, err)

	assert.Equal(t, event.PackageID, receivedEvent.PackageID)
	assert.Equal(t, event.CourierID, receivedEvent.CourierID)
	assert.Equal(t, event.PickupLocation.Latitude, receivedEvent.PickupLocation.Latitude)
	assert.Equal(t, event.PickupLocation.Longitude, receivedEvent.PickupLocation.Longitude)

	// Verify partition key
	assert.Equal(t, event.CourierID, messages[0].Metadata.Get("partition_key"))
}

func TestStatusPublisher_PublishDelivery(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	ctx := context.Background()

	event := DeliverOrderEvent{
		PackageID: "pkg-123",
		CourierID: "courier-456",
		Status:    DeliveryStatusDelivered,
		CurrentLocation: Location{
			Latitude:  52.5300,
			Longitude: 13.4150,
			Accuracy:  10.0,
			Timestamp: time.Now(),
		},
		DeliveredAt: time.Now(),
	}

	err := statusPub.PublishDelivery(ctx, event)
	require.NoError(t, err)

	// Verify message was published
	messages := mockPub.messages[TopicDeliverOrder]
	require.Len(t, messages, 1)

	// Verify message content
	var receivedEvent DeliverOrderEvent
	err = json.Unmarshal(messages[0].Payload, &receivedEvent)
	require.NoError(t, err)

	assert.Equal(t, event.PackageID, receivedEvent.PackageID)
	assert.Equal(t, event.CourierID, receivedEvent.CourierID)
	assert.Equal(t, DeliveryStatusDelivered, receivedEvent.Status)

	// Verify partition key
	assert.Equal(t, event.CourierID, messages[0].Metadata.Get("partition_key"))
}

func TestStatusPublisher_PublishDeliveryNotDelivered(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	ctx := context.Background()

	event := DeliverOrderEvent{
		PackageID: "pkg-789",
		CourierID: "courier-101",
		Status:    DeliveryStatusNotDelivered,
		Reason:    ReasonCustomerNotAvailable,
		CurrentLocation: Location{
			Latitude:  52.5400,
			Longitude: 13.4200,
			Accuracy:  15.0,
			Timestamp: time.Now(),
		},
		DeliveredAt: time.Now(),
	}

	err := statusPub.PublishDelivery(ctx, event)
	require.NoError(t, err)

	messages := mockPub.messages[TopicDeliverOrder]
	require.Len(t, messages, 1)

	var receivedEvent DeliverOrderEvent
	err = json.Unmarshal(messages[0].Payload, &receivedEvent)
	require.NoError(t, err)

	assert.Equal(t, DeliveryStatusNotDelivered, receivedEvent.Status)
	assert.Equal(t, ReasonCustomerNotAvailable, receivedEvent.Reason)
}

func TestNewPickUpOrderEvent(t *testing.T) {
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())
	location := vo.MustNewLocation(52.5200, 13.4050)

	event := NewPickUpOrderEvent("courier-1", order, location)

	assert.Equal(t, "pkg-1", event.PackageID)
	assert.Equal(t, "courier-1", event.CourierID)
	assert.Equal(t, location.Latitude(), event.PickupLocation.Latitude)
	assert.Equal(t, location.Longitude(), event.PickupLocation.Longitude)
	assert.NotZero(t, event.PickedUpAt)
}

func TestNewDeliverOrderEvent_Delivered(t *testing.T) {
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())
	location := vo.MustNewLocation(52.5300, 13.4150)

	event := NewDeliverOrderEvent("courier-1", order, location, true, "")

	assert.Equal(t, "pkg-1", event.PackageID)
	assert.Equal(t, "courier-1", event.CourierID)
	assert.Equal(t, DeliveryStatusDelivered, event.Status)
	assert.Empty(t, event.Reason)
	assert.NotZero(t, event.DeliveredAt)
}

func TestNewDeliverOrderEvent_NotDelivered(t *testing.T) {
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())
	location := vo.MustNewLocation(52.5300, 13.4150)

	event := NewDeliverOrderEvent("courier-1", order, location, false, ReasonCustomerRefused)

	assert.Equal(t, DeliveryStatusNotDelivered, event.Status)
	assert.Equal(t, ReasonCustomerRefused, event.Reason)
}

func TestStatusPublisher_Close(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	err := statusPub.Close()
	require.NoError(t, err)
	assert.True(t, mockPub.closed)
}

// Ensure topic constants are correct
func TestTopicConstants(t *testing.T) {
	assert.Equal(t, "delivery.command.pick_up_order", TopicPickUpOrder)
	assert.Equal(t, "delivery.command.deliver_order", TopicDeliverOrder)
	assert.Equal(t, "courier.location.updates", TopicCourierLocation)
	assert.Equal(t, "delivery.order.assigned", TopicOrderAssigned)
}

// Ensure status constants are correct
func TestDeliveryStatusConstants(t *testing.T) {
	assert.Equal(t, "DELIVERED", DeliveryStatusDelivered)
	assert.Equal(t, "NOT_DELIVERED", DeliveryStatusNotDelivered)
}

// Ensure reason constants are correct
func TestReasonConstants(t *testing.T) {
	assert.Equal(t, "CUSTOMER_NOT_AVAILABLE", ReasonCustomerNotAvailable)
	assert.Equal(t, "WRONG_ADDRESS", ReasonWrongAddress)
	assert.Equal(t, "CUSTOMER_REFUSED", ReasonCustomerRefused)
	assert.Equal(t, "ACCESS_DENIED", ReasonAccessDenied)
	assert.Equal(t, "PACKAGE_DAMAGED", ReasonPackageDamaged)
	assert.Equal(t, "OTHER", ReasonOther)
}
