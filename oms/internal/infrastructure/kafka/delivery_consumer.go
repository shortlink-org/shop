package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	deliverycommon "github.com/shortlink-org/shop/oms/internal/domain/delivery/common/v1"
	deliveryevents "github.com/shortlink-org/shop/oms/internal/domain/delivery/events/v1"
)

// DeliveryEventType represents the type of delivery event.
type DeliveryEventType string

const (
	EventTypePackageAccepted     DeliveryEventType = "PACKAGE_ACCEPTED"
	EventTypePackageAssigned     DeliveryEventType = "PACKAGE_ASSIGNED"
	EventTypePackageInTransit    DeliveryEventType = "PACKAGE_IN_TRANSIT"
	EventTypePackageDelivered    DeliveryEventType = "PACKAGE_DELIVERED"
	EventTypePackageNotDelivered DeliveryEventType = "PACKAGE_NOT_DELIVERED"

	// ConsumerGroupOMSDelivery is the consumer group for OMS delivery events.
	ConsumerGroupOMSDelivery = "oms-delivery-consumer"

	// TopicDeliveryPackageStatus is the topic for delivery package status events.
	TopicDeliveryPackageStatus = "delivery.package.status.v1"
)

var (
	errUnsupportedEventType = errors.New("unsupported or non-status event_type")
	errConsumerClosed       = errors.New("consumer closed")
)

// DeliveryStatusEvent represents a delivery status update from the Delivery service.
type DeliveryStatusEvent struct {
	MessageID           string                              `json:"-"`
	PackageID           uuid.UUID                           `json:"package_id"`
	OrderID             uuid.UUID                           `json:"order_id"`
	CourierID           uuid.UUID                           `json:"courier_id"`
	Status              string                              `json:"status"`
	EventType           DeliveryEventType                   `json:"event_type"`
	Reason              string                              `json:"reason,omitempty"`
	Description         string                              `json:"reason_description,omitempty"`
	OccurredAt          time.Time                           `json:"occurred_at"`
	NotDeliveredDetails *deliverycommon.NotDeliveredDetails `json:"-"`
	DeliveryLocation    *deliverycommon.Location            `json:"-"`
}

// DeliveryEventHandler defines the interface for handling delivery events.
type DeliveryEventHandler interface {
	HandleDeliveryStatus(ctx context.Context, event DeliveryStatusEvent) error
}

// DeliveryConsumer consumes delivery status events from Kafka using Watermill.
type DeliveryConsumer struct {
	topic      string
	handler    DeliveryEventHandler
	log        logger.Logger
	subscriber message.Subscriber
	cancel     context.CancelCauseFunc
}

// NewDeliveryConsumer creates a new delivery consumer.
func NewDeliveryConsumer(
	topic string,
	subscriber message.Subscriber,
	handler DeliveryEventHandler,
	log logger.Logger,
) *DeliveryConsumer {
	return &DeliveryConsumer{
		topic:      topic,
		handler:    handler,
		log:        log,
		subscriber: subscriber,
	}
}

// Start starts consuming messages in a goroutine.
func (c *DeliveryConsumer) Start(ctx context.Context) error {
	c.log.Info("Starting delivery consumer",
		slog.String("topic", c.topic))

	messages, err := c.subscriber.Subscribe(ctx, c.topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	ctx, c.cancel = context.WithCancelCause(ctx)

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
//
//nolint:funcorder // unexported handler
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

	event.MessageID = msg.UUID

	err = c.handler.HandleDeliveryStatus(ctx, event)
	if err != nil {
		c.log.Error("Failed to handle delivery event",
			slog.Any("error", err),
			slog.String("package_id", event.PackageID.String()),
			slog.String("order_id", event.OrderID.String()))
		msg.Nack()

		return
	}

	msg.Ack()
	c.log.Info("Processed delivery event",
		slog.String("package_id", event.PackageID.String()),
		slog.String("order_id", event.OrderID.String()),
		slog.String("status", event.Status))
}

// unmarshalDeliveryEvent decodes payload by event_type and maps to DeliveryStatusEvent.
//
//nolint:funcorder // unexported helper
func (c *DeliveryConsumer) unmarshalDeliveryEvent(eventType string, payload []byte) (DeliveryStatusEvent, error) {
	var event DeliveryStatusEvent

	switch eventType {
	case "PackageAcceptedEvent":
		var e deliveryevents.PackageAcceptedEvent

		err := proto.Unmarshal(payload, &e)
		if err != nil {
			return event, fmt.Errorf("proto unmarshal PackageAcceptedEvent: %w", err)
		}

		event, err = mapAcceptedToStatusEvent(&e)
		if err != nil {
			return event, err
		}
	case "PackageAssignedEvent":
		var e deliveryevents.PackageAssignedEvent

		err := proto.Unmarshal(payload, &e)
		if err != nil {
			return event, fmt.Errorf("proto unmarshal PackageAssignedEvent: %w", err)
		}

		event, err = mapAssignedToStatusEvent(&e)
		if err != nil {
			return event, err
		}
	case "PackageInTransitEvent":
		var e deliveryevents.PackageInTransitEvent

		err := proto.Unmarshal(payload, &e)
		if err != nil {
			return event, fmt.Errorf("proto unmarshal PackageInTransitEvent: %w", err)
		}

		event, err = mapInTransitToStatusEvent(&e)
		if err != nil {
			return event, err
		}
	case "PackageDeliveredEvent":
		var e deliveryevents.PackageDeliveredEvent

		err := proto.Unmarshal(payload, &e)
		if err != nil {
			return event, fmt.Errorf("proto unmarshal PackageDeliveredEvent: %w", err)
		}

		event, err = mapDeliveredToStatusEvent(&e)
		if err != nil {
			return event, err
		}
	case "PackageNotDeliveredEvent":
		var e deliveryevents.PackageNotDeliveredEvent

		err := proto.Unmarshal(payload, &e)
		if err != nil {
			return event, fmt.Errorf("proto unmarshal PackageNotDeliveredEvent: %w", err)
		}

		event, err = mapNotDeliveredToStatusEvent(&e)
		if err != nil {
			return event, err
		}
	default:
		// Ignore other event types.
		return event, fmt.Errorf("unsupported or non-status event_type %s: %w", eventType, errUnsupportedEventType)
	}

	return event, nil
}

func mapAcceptedToStatusEvent(e *deliveryevents.PackageAcceptedEvent) (DeliveryStatusEvent, error) {
	packageID, err := parseRequiredUUID(e.GetPackageId(), "package_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	orderID, err := parseRequiredUUID(e.GetOrderId(), "order_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	return DeliveryStatusEvent{
		PackageID:  packageID,
		OrderID:    orderID,
		Status:     deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:  EventTypePackageAccepted,
		OccurredAt: timestampToTime(e.GetOccurredAt()),
	}, nil
}

func mapAssignedToStatusEvent(e *deliveryevents.PackageAssignedEvent) (DeliveryStatusEvent, error) {
	packageID, err := parseRequiredUUID(e.GetPackageId(), "package_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	courierID, err := parseRequiredUUID(e.GetCourierId(), "courier_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	return DeliveryStatusEvent{
		PackageID:  packageID,
		CourierID:  courierID,
		Status:     deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:  EventTypePackageAssigned,
		OccurredAt: timestampToTime(e.GetOccurredAt()),
	}, nil
}

func mapInTransitToStatusEvent(e *deliveryevents.PackageInTransitEvent) (DeliveryStatusEvent, error) {
	packageID, err := parseRequiredUUID(e.GetPackageId(), "package_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	orderID, err := parseOptionalUUID(e.GetOrderId(), "order_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	courierID, err := parseRequiredUUID(e.GetCourierId(), "courier_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	return DeliveryStatusEvent{
		PackageID:  packageID,
		OrderID:    orderID,
		CourierID:  courierID,
		Status:     deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:  EventTypePackageInTransit,
		OccurredAt: timestampToTime(e.GetOccurredAt()),
	}, nil
}

func mapDeliveredToStatusEvent(e *deliveryevents.PackageDeliveredEvent) (DeliveryStatusEvent, error) {
	packageID, err := parseRequiredUUID(e.GetPackageId(), "package_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	orderID, err := parseOptionalUUID(e.GetOrderId(), "order_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	courierID, err := parseRequiredUUID(e.GetCourierId(), "courier_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	return DeliveryStatusEvent{
		PackageID:        packageID,
		OrderID:          orderID,
		CourierID:        courierID,
		Status:           deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:        EventTypePackageDelivered,
		OccurredAt:       timestampToTime(e.GetOccurredAt()),
		DeliveryLocation: e.GetDeliveryLocation(),
	}, nil
}

func mapNotDeliveredToStatusEvent(e *deliveryevents.PackageNotDeliveredEvent) (DeliveryStatusEvent, error) {
	var reasonStr, desc string
	if d := e.GetNotDeliveredDetails(); d != nil {
		reasonStr = deliverycommon.NotDeliveredReason_name[int32(d.GetReason())]
		desc = d.GetDescription()
	}

	packageID, err := parseRequiredUUID(e.GetPackageId(), "package_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	orderID, err := parseOptionalUUID(e.GetOrderId(), "order_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	courierID, err := parseRequiredUUID(e.GetCourierId(), "courier_id")
	if err != nil {
		return DeliveryStatusEvent{}, err
	}

	return DeliveryStatusEvent{
		PackageID:           packageID,
		OrderID:             orderID,
		CourierID:           courierID,
		Status:              deliverycommon.PackageStatus_name[int32(e.GetStatus())],
		EventType:           EventTypePackageNotDelivered,
		Reason:              reasonStr,
		Description:         desc,
		OccurredAt:          timestampToTime(e.GetOccurredAt()),
		NotDeliveredDetails: e.GetNotDeliveredDetails(),
	}, nil
}

func parseRequiredUUID(value, fieldName string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil, fmt.Errorf("%s is required", fieldName)
	}

	return parseOptionalUUID(value, fieldName)
}

func parseOptionalUUID(value, fieldName string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s: %w", fieldName, err)
	}

	return parsed, nil
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
		c.cancel(errConsumerClosed)
	}

	err := c.subscriber.Close()
	if err != nil {
		return fmt.Errorf("close subscriber: %w", err)
	}

	return nil
}
