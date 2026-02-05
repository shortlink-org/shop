package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

const (
	// TopicCourierLocation is the Kafka topic for courier location events.
	// Format: {domain}.{entity}.{event}.v1
	TopicCourierLocation = "delivery.courier.location_received.v1"
)

// LocationPublisher publishes courier location events to Kafka.
type LocationPublisher struct {
	publisher message.Publisher
}

// NewLocationPublisher creates a new Kafka location publisher using go-sdk/watermill publisher.
func NewLocationPublisher(publisher message.Publisher) *LocationPublisher {
	return &LocationPublisher{
		publisher: publisher,
	}
}

// PublishLocation publishes a courier location event to Kafka.
func (p *LocationPublisher) PublishLocation(ctx context.Context, event vo.CourierLocationEvent) error {
	payload, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("event to json: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	// Set partition key to courier ID for ordered delivery per courier
	msg.Metadata.Set("partition_key", event.CourierID)

	if err := p.publisher.Publish(TopicCourierLocation, msg); err != nil {
		return fmt.Errorf("publish location: %w", err)
	}

	return nil
}

// PublishLocationBatch publishes multiple courier location events.
func (p *LocationPublisher) PublishLocationBatch(ctx context.Context, events []vo.CourierLocationEvent) error {
	messages := make([]*message.Message, 0, len(events))

	for _, event := range events {
		payload, err := event.ToJSON()
		if err != nil {
			return fmt.Errorf("event to json: %w", err)
		}

		msg := message.NewMessage(watermill.NewUUID(), payload)
		msg.Metadata.Set("partition_key", event.CourierID)
		messages = append(messages, msg)
	}

	err := p.publisher.Publish(TopicCourierLocation, messages...)
	if err != nil {
		return fmt.Errorf("publish location batch: %w", err)
	}

	return nil
}

// Close closes the publisher.
func (p *LocationPublisher) Close() error {
	err := p.publisher.Close()
	if err != nil {
		return fmt.Errorf("publisher close: %w", err)
	}

	return nil
}

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

// PublishPickUp publishes an order picked up event.
func (p *KafkaStatusPublisher) PublishPickUp(ctx context.Context, event PickUpOrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal pickup event: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	// Partition by order so lifecycle order is preserved (handover-safe).
	msg.Metadata.Set(metadataKeyPartitionKey, event.OrderID)

	if err := p.publisher.Publish(TopicPickUpOrder, msg); err != nil {
		return fmt.Errorf("publish pickup: %w", err)
	}

	return nil
}

// PublishDelivery publishes an order delivered event.
func (p *KafkaStatusPublisher) PublishDelivery(ctx context.Context, event DeliverOrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal delivery event: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	// Partition by order so lifecycle order is preserved (handover-safe).
	msg.Metadata.Set(metadataKeyPartitionKey, event.OrderID)

	if err := p.publisher.Publish(TopicDeliverOrder, msg); err != nil {
		return fmt.Errorf("publish delivery: %w", err)
	}

	return nil
}

// Close closes the status publisher.
func (p *KafkaStatusPublisher) Close() error {
	err := p.publisher.Close()
	if err != nil {
		return fmt.Errorf("status publisher close: %w", err)
	}

	return nil
}
