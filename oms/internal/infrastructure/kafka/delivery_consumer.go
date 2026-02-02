package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/shortlink-org/go-sdk/logger"
)

// DeliveryEventType represents the type of delivery event.
type DeliveryEventType string

const (
	EventTypePackageInTransit    DeliveryEventType = "PACKAGE_IN_TRANSIT"
	EventTypePackageDelivered    DeliveryEventType = "PACKAGE_DELIVERED"
	EventTypePackageNotDelivered DeliveryEventType = "PACKAGE_NOT_DELIVERED"

	// ConsumerGroupOMSDelivery is the consumer group for OMS delivery events.
	ConsumerGroupOMSDelivery = "oms-delivery-consumer"

	// TopicDeliveryPackageStatus is the topic for delivery package status events.
	TopicDeliveryPackageStatus = "delivery.package.status.v1"
)

// DeliveryStatusEvent represents a delivery status update from the Delivery service.
type DeliveryStatusEvent struct {
	PackageID   string            `json:"package_id"`
	OrderID     string            `json:"order_id"`
	CourierID   string            `json:"courier_id"`
	Status      string            `json:"status"`
	EventType   DeliveryEventType `json:"event_type"`
	Reason      string            `json:"reason,omitempty"`
	Description string            `json:"reason_description,omitempty"`
	OccurredAt  time.Time         `json:"occurred_at"`
}

// DeliveryEventHandler defines the interface for handling delivery events.
type DeliveryEventHandler interface {
	HandleDeliveryStatus(ctx context.Context, event DeliveryStatusEvent) error
}

// DeliveryConsumerConfig contains configuration for the delivery consumer.
type DeliveryConsumerConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

// DefaultDeliveryConsumerConfig returns default configuration.
func DefaultDeliveryConsumerConfig() DeliveryConsumerConfig {
	return DeliveryConsumerConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: ConsumerGroupOMSDelivery,
		Topic:   TopicDeliveryPackageStatus,
	}
}

// watermillLoggerAdapter adapts go-sdk logger to Watermill logger interface.
type watermillLoggerAdapter struct {
	log logger.Logger
}

func (w *watermillLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	w.log.Error(fmt.Sprintf("%s: %v", msg, err), slog.String("error", err.Error()))
}

func (w *watermillLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	w.log.Info(msg)
}

func (w *watermillLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	w.log.Debug(msg)
}

func (w *watermillLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	w.log.Debug(msg)
}

func (w *watermillLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return w
}

// DeliveryConsumer consumes delivery status events from Kafka using Watermill.
type DeliveryConsumer struct {
	config     DeliveryConsumerConfig
	handler    DeliveryEventHandler
	log        logger.Logger
	subscriber *kafka.Subscriber
	cancel     context.CancelFunc
}

// NewDeliveryConsumer creates a new delivery consumer.
func NewDeliveryConsumer(
	config DeliveryConsumerConfig,
	handler DeliveryEventHandler,
	log logger.Logger,
) (*DeliveryConsumer, error) {
	saramaConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	wmLogger := &watermillLoggerAdapter{log: log}

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               config.Brokers,
			Unmarshaler:           kafka.DefaultMarshaler{},
			ConsumerGroup:         config.GroupID,
			OverwriteSaramaConfig: saramaConfig,
		},
		wmLogger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka subscriber: %w", err)
	}

	return &DeliveryConsumer{
		config:     config,
		handler:    handler,
		log:        log,
		subscriber: subscriber,
	}, nil
}

// Start starts consuming messages in a goroutine.
func (c *DeliveryConsumer) Start(ctx context.Context) error {
	c.log.Info("Starting delivery consumer",
		slog.String("topic", c.config.Topic),
		slog.String("group_id", c.config.GroupID))

	messages, err := c.subscriber.Subscribe(ctx, c.config.Topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	ctx, c.cancel = context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-messages:
				if msg == nil {
					continue
				}
				c.processMessage(ctx, msg)
			}
		}
	}()

	c.log.Info("Delivery consumer started")
	return nil
}

// processMessage processes a single message.
func (c *DeliveryConsumer) processMessage(ctx context.Context, msg *message.Message) {
	c.log.Debug("Received message",
		slog.String("uuid", msg.UUID))

	var event DeliveryStatusEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		c.log.Error("Failed to unmarshal delivery event",
			slog.Any("error", err),
			slog.String("payload", string(msg.Payload)))
		msg.Ack()
		return
	}

	// Determine event type from status if not set
	if event.EventType == "" {
		event.EventType = statusToEventType(event.Status)
	}

	if err := c.handler.HandleDeliveryStatus(ctx, event); err != nil {
		c.log.Error("Failed to handle delivery event",
			slog.Any("error", err),
			slog.String("package_id", event.PackageID),
			slog.String("order_id", event.OrderID))
		msg.Nack()
		return
	}

	msg.Ack()
	c.log.Info("Processed delivery event",
		slog.String("package_id", event.PackageID),
		slog.String("order_id", event.OrderID),
		slog.String("status", event.Status))
}

// Close closes the consumer.
func (c *DeliveryConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.subscriber.Close()
}

// statusToEventType converts status string to event type.
func statusToEventType(status string) DeliveryEventType {
	switch status {
	case "IN_TRANSIT":
		return EventTypePackageInTransit
	case "DELIVERED":
		return EventTypePackageDelivered
	case "NOT_DELIVERED":
		return EventTypePackageNotDelivered
	default:
		return ""
	}
}
