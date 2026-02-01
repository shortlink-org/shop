package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

const (
	// TopicPickUpOrder is the Kafka topic for pickup order commands.
	TopicPickUpOrder = "delivery.command.pick_up_order"
	// TopicDeliverOrder is the Kafka topic for deliver order commands.
	TopicDeliverOrder = "delivery.command.deliver_order"
)

// PickUpOrderEvent represents a pickup order command event.
type PickUpOrderEvent struct {
	PackageID      string    `json:"package_id"`
	CourierID      string    `json:"courier_id"`
	PickupLocation Location  `json:"pickup_location"`
	PickedUpAt     time.Time `json:"picked_up_at"`
}

// DeliverOrderEvent represents a deliver order command event.
type DeliverOrderEvent struct {
	PackageID       string    `json:"package_id"`
	CourierID       string    `json:"courier_id"`
	Status          string    `json:"status"` // DELIVERED or NOT_DELIVERED
	Reason          string    `json:"reason,omitempty"`
	CurrentLocation Location  `json:"current_location"`
	DeliveredAt     time.Time `json:"delivered_at"`
}

// Location represents a geographic location in events.
type Location struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy"`
	Timestamp time.Time `json:"timestamp"`
}

// DeliveryStatus constants for delivery outcome.
const (
	DeliveryStatusDelivered    = "DELIVERED"
	DeliveryStatusNotDelivered = "NOT_DELIVERED"
)

// NotDeliveredReason constants for failed delivery reasons.
const (
	ReasonCustomerNotAvailable = "CUSTOMER_NOT_AVAILABLE"
	ReasonWrongAddress         = "WRONG_ADDRESS"
	ReasonCustomerRefused      = "CUSTOMER_REFUSED"
	ReasonAccessDenied         = "ACCESS_DENIED"
	ReasonPackageDamaged       = "PACKAGE_DAMAGED"
	ReasonOther                = "OTHER"
)

// StatusPublisher defines the interface for publishing delivery status events.
type StatusPublisher interface {
	PublishPickUp(ctx context.Context, event PickUpOrderEvent) error
	PublishDelivery(ctx context.Context, event DeliverOrderEvent) error
	Close() error
}

// KafkaStatusPublisher publishes delivery status events to Kafka.
type KafkaStatusPublisher struct {
	publisher message.Publisher
}

// NewStatusPublisher creates a new Kafka status publisher.
func NewStatusPublisher(publisher message.Publisher) *KafkaStatusPublisher {
	return &KafkaStatusPublisher{
		publisher: publisher,
	}
}

// PublishPickUp publishes a pickup order event.
func (p *KafkaStatusPublisher) PublishPickUp(ctx context.Context, event PickUpOrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("partition_key", event.CourierID)

	return p.publisher.Publish(TopicPickUpOrder, msg)
}

// PublishDelivery publishes a deliver order event.
func (p *KafkaStatusPublisher) PublishDelivery(ctx context.Context, event DeliverOrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("partition_key", event.CourierID)

	return p.publisher.Publish(TopicDeliverOrder, msg)
}

// Close closes the publisher.
func (p *KafkaStatusPublisher) Close() error {
	return p.publisher.Close()
}

// NewPickUpOrderEvent creates a pickup order event from domain objects.
func NewPickUpOrderEvent(courierID string, order vo.DeliveryOrder, location vo.Location) PickUpOrderEvent {
	return PickUpOrderEvent{
		PackageID: order.PackageID(),
		CourierID: courierID,
		PickupLocation: Location{
			Latitude:  location.Latitude(),
			Longitude: location.Longitude(),
			Accuracy:  10.0, // Default accuracy
			Timestamp: time.Now(),
		},
		PickedUpAt: time.Now(),
	}
}

// NewDeliverOrderEvent creates a deliver order event from domain objects.
func NewDeliverOrderEvent(
	courierID string,
	order vo.DeliveryOrder,
	location vo.Location,
	delivered bool,
	reason string,
) DeliverOrderEvent {
	status := DeliveryStatusDelivered
	if !delivered {
		status = DeliveryStatusNotDelivered
	}

	return DeliverOrderEvent{
		PackageID: order.PackageID(),
		CourierID: courierID,
		Status:    status,
		Reason:    reason,
		CurrentLocation: Location{
			Latitude:  location.Latitude(),
			Longitude: location.Longitude(),
			Accuracy:  10.0,
			Timestamp: time.Now(),
		},
		DeliveredAt: time.Now(),
	}
}
