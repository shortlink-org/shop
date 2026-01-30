package kafka

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

const (
	// TopicCourierLocation is the Kafka topic for courier location events.
	TopicCourierLocation = "courier.location.updates"
)

// PublisherConfig holds configuration for the Kafka publisher.
type PublisherConfig struct {
	Brokers []string
}

// DefaultPublisherConfig returns default configuration.
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		Brokers: []string{"localhost:9092"},
	}
}

// LocationPublisher publishes courier location events to Kafka.
type LocationPublisher struct {
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

// NewLocationPublisher creates a new Kafka location publisher.
func NewLocationPublisher(config PublisherConfig, logger watermill.LoggerAdapter) (*LocationPublisher, error) {
	if logger == nil {
		logger = watermill.NewStdLogger(false, false)
	}

	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   config.Brokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &LocationPublisher{
		publisher: publisher,
		logger:    logger,
	}, nil
}

// PublishLocation publishes a courier location event to Kafka.
func (p *LocationPublisher) PublishLocation(ctx context.Context, event vo.CourierLocationEvent) error {
	payload, err := event.ToJSON()
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	// Set partition key to courier ID for ordered delivery per courier
	msg.Metadata.Set("partition_key", event.CourierID)

	return p.publisher.Publish(TopicCourierLocation, msg)
}

// PublishLocationBatch publishes multiple courier location events.
func (p *LocationPublisher) PublishLocationBatch(ctx context.Context, events []vo.CourierLocationEvent) error {
	messages := make([]*message.Message, 0, len(events))

	for _, event := range events {
		payload, err := event.ToJSON()
		if err != nil {
			return err
		}

		msg := message.NewMessage(watermill.NewUUID(), payload)
		msg.Metadata.Set("partition_key", event.CourierID)
		messages = append(messages, msg)
	}

	return p.publisher.Publish(TopicCourierLocation, messages...)
}

// Close closes the publisher.
func (p *LocationPublisher) Close() error {
	return p.publisher.Close()
}
