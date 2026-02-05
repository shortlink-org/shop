package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusPublisher_PublishPickUp(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	ctx := context.Background()

	event := PickUpOrderEvent{
		OrderID:   "order-123",
		CourierID: "courier-456",
		PickupLocation: Location{
			Latitude:  52.5200,
			Longitude: 13.4050,
			Accuracy:  defaultLocationAccuracy,
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

	assert.Equal(t, event.OrderID, receivedEvent.OrderID)
	assert.Equal(t, event.CourierID, receivedEvent.CourierID)
	assert.Equal(t, event.PickupLocation.Latitude, receivedEvent.PickupLocation.Latitude)
	assert.Equal(t, event.PickupLocation.Longitude, receivedEvent.PickupLocation.Longitude)

	// Verify partition key (by order for lifecycle ordering)
	assert.Equal(t, event.OrderID, messages[0].Metadata.Get("partition_key"))
}

func TestStatusPublisher_PublishDelivery(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	ctx := context.Background()

	event := DeliverOrderEvent{
		OrderID:   "order-123",
		CourierID: "courier-456",
		Status:    DeliveryStatusDelivered,
		CurrentLocation: Location{
			Latitude:  52.5300,
			Longitude: 13.4150,
			Accuracy:  defaultLocationAccuracy,
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

	assert.Equal(t, event.OrderID, receivedEvent.OrderID)
	assert.Equal(t, event.CourierID, receivedEvent.CourierID)
	assert.Equal(t, DeliveryStatusDelivered, receivedEvent.Status)

	// Verify partition key (by order for lifecycle ordering)
	assert.Equal(t, event.OrderID, messages[0].Metadata.Get("partition_key"))
}

func TestStatusPublisher_PublishDeliveryNotDelivered(t *testing.T) {
	mockPub := newMockPublisher()
	statusPub := NewStatusPublisher(mockPub)

	ctx := context.Background()

	event := DeliverOrderEvent{
		OrderID:   "order-789",
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

	assert.Equal(t, "order-1", event.OrderID)
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

	event, err := NewDeliverOrderEvent("courier-1", order, location, true, "")
	require.NoError(t, err)

	assert.Equal(t, "order-1", event.OrderID)
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

	event, err := NewDeliverOrderEvent("courier-1", order, location, false, ReasonCustomerRefused)
	require.NoError(t, err)

	assert.Equal(t, "order-1", event.OrderID)
	assert.Equal(t, DeliveryStatusNotDelivered, event.Status)
	assert.Equal(t, ReasonCustomerRefused, event.Reason)
}

func TestNewDeliverOrderEvent_Validation(t *testing.T) {
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())
	location := vo.MustNewLocation(52.5300, 13.4150)

	t.Run("delivered_with_reason_returns_error", func(t *testing.T) {
		_, err := NewDeliverOrderEvent("c1", order, location, true, ReasonCustomerRefused)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrReasonMustBeEmpty))
		assert.Contains(t, err.Error(), "got=\"CUSTOMER_REFUSED\"")
	})

	t.Run("not_delivered_empty_reason_returns_error", func(t *testing.T) {
		_, err := NewDeliverOrderEvent("c1", order, location, false, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrReasonRequired))
	})

	t.Run("not_delivered_invalid_reason_returns_error", func(t *testing.T) {
		_, err := NewDeliverOrderEvent("c1", order, location, false, NotDeliveredReason("INVALID_REASON"))
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidReason))
		assert.Contains(t, err.Error(), "got=\"INVALID_REASON\"")
	})

	t.Run("not_delivered_OTHER_valid", func(t *testing.T) {
		event, err := NewDeliverOrderEvent("c1", order, location, false, ReasonOther)
		require.NoError(t, err)
		assert.Equal(t, DeliveryStatusNotDelivered, event.Status)
		assert.Equal(t, ReasonOther, event.Reason)
	})
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
	assert.Equal(t, "delivery.order.order_picked_up.v1", TopicPickUpOrder)
	assert.Equal(t, "delivery.order.order_delivered.v1", TopicDeliverOrder)
	assert.Equal(t, "delivery.courier.location_received.v1", TopicCourierLocation)
	assert.Equal(t, "delivery.order.assigned.v1", TopicOrderAssigned)
}

// Ensure status constants serialize correctly
func TestDeliveryStatusConstants(t *testing.T) {
	assert.Equal(t, "DELIVERED", string(DeliveryStatusDelivered))
	assert.Equal(t, "NOT_DELIVERED", string(DeliveryStatusNotDelivered))
}

// Ensure reason constants serialize correctly
func TestReasonConstants(t *testing.T) {
	assert.Equal(t, "CUSTOMER_NOT_AVAILABLE", string(ReasonCustomerNotAvailable))
	assert.Equal(t, "WRONG_ADDRESS", string(ReasonWrongAddress))
	assert.Equal(t, "CUSTOMER_REFUSED", string(ReasonCustomerRefused))
	assert.Equal(t, "ACCESS_DENIED", string(ReasonAccessDenied))
	assert.Equal(t, "PACKAGE_DAMAGED", string(ReasonPackageDamaged))
	assert.Equal(t, "OTHER", string(ReasonOther))
}
