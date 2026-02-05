package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shortlink-org/go-sdk/logger"
	deliverycommon "github.com/shortlink-org/shop/oms/internal/domain/delivery/common/v1"
	deliveryevents "github.com/shortlink-org/shop/oms/internal/domain/delivery/events/v1"
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

// Kafka header set by Delivery service for event type dispatch.
const eventTypeHeader = "event_type"

// processMessage processes a single message (protobuf-encoded with event_type header).
func (c *DeliveryConsumer) processMessage(ctx context.Context, msg *message.Message) {
	c.log.Debug("Received message",
		slog.String("uuid", msg.UUID))

	eventType := msg.Metadata.Get(eventTypeHeader)
	if eventType == "" {
		c.log.Error("Missing event_type header, cannot decode delivery event",
			slog.String("uuid", msg.UUID))
		msg.Ack()
		return
	}

	event, err := c.unmarshalDeliveryEvent(eventType, msg.Payload)
	if err != nil {
		c.log.Error("Failed to unmarshal delivery event",
			slog.Any("error", err),
			slog.String("event_type", eventType))
		msg.Ack()
		return
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

// unmarshalDeliveryEvent decodes payload by event_type and maps to DeliveryStatusEvent.
func (c *DeliveryConsumer) unmarshalDeliveryEvent(eventType string, payload []byte) (DeliveryStatusEvent, error) {
	var event DeliveryStatusEvent
	switch eventType {
	case "PackageInTransitEvent":
		var e deliveryevents.PackageInTransitEvent
		if err := proto.Unmarshal(payload, &e); err != nil {
			return event, fmt.Errorf("proto unmarshal PackageInTransitEvent: %w", err)
		}
		event = mapInTransitToStatusEvent(&e)
	case "PackageDeliveredEvent":
		var e deliveryevents.PackageDeliveredEvent
		if err := proto.Unmarshal(payload, &e); err != nil {
			return event, fmt.Errorf("proto unmarshal PackageDeliveredEvent: %w", err)
		}
		event = mapDeliveredToStatusEvent(&e)
	case "PackageNotDeliveredEvent":
		var e deliveryevents.PackageNotDeliveredEvent
		if err := proto.Unmarshal(payload, &e); err != nil {
			return event, fmt.Errorf("proto unmarshal PackageNotDeliveredEvent: %w", err)
		}
		event = mapNotDeliveredToStatusEvent(&e)
	default:
		// Ignore other event types (PackageAcceptedEvent, PackageAssignedEvent, etc.)
		return event, fmt.Errorf("unsupported or non-status event_type: %s", eventType)
	}
	return event, nil
}

func mapInTransitToStatusEvent(e *deliveryevents.PackageInTransitEvent) DeliveryStatusEvent {
	return DeliveryStatusEvent{
		PackageID:  e.GetPackageId(),
		OrderID:    e.GetOrderId(),
		CourierID:  e.GetCourierId(),
		Status:     deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:  EventTypePackageInTransit,
		OccurredAt: timestampToTime(e.GetOccurredAt()),
	}
}

func mapDeliveredToStatusEvent(e *deliveryevents.PackageDeliveredEvent) DeliveryStatusEvent {
	return DeliveryStatusEvent{
		PackageID:  e.GetPackageId(),
		OrderID:    e.GetOrderId(),
		CourierID:  e.GetCourierId(),
		Status:     deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:  EventTypePackageDelivered,
		OccurredAt: timestampToTime(e.GetOccurredAt()),
	}
}

func mapNotDeliveredToStatusEvent(e *deliveryevents.PackageNotDeliveredEvent) DeliveryStatusEvent {
	reasonStr := deliverycommon.NotDeliveredReason_name[int32(e.GetReason())]
	return DeliveryStatusEvent{
		PackageID:   e.GetPackageId(),
		OrderID:     e.GetOrderId(),
		CourierID:   e.GetCourierId(),
		Status:      deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:   EventTypePackageNotDelivered,
		Reason:      reasonStr,
		Description: e.GetReasonDescription(),
		OccurredAt:  timestampToTime(e.GetOccurredAt()),
	}
}

func timestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// Close closes the consumer.
func (c *DeliveryConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.subscriber.Close()
}
