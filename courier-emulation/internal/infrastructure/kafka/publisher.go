package kafka

import (
	"context"
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
