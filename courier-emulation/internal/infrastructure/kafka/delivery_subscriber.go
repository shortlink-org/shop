package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

const (
	// TopicOrderAssigned is the Kafka topic for order assignment events from Delivery Service.
	// Format: {domain}.{entity}.{event}.v1
	TopicOrderAssigned = "delivery.order.assigned.v1"
	// ConsumerGroupCourierEmulation is the consumer group for this service.
	ConsumerGroupCourierEmulation = "courier-emulation"
)

// Address represents a delivery address with location coordinates.
// Matches proto: domain.delivery.common.v1.Address
type Address struct {
	Street     string  `json:"street,omitempty"`
	City       string  `json:"city,omitempty"`
	PostalCode string  `json:"postal_code,omitempty"`
	Country    string  `json:"country,omitempty"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

// DeliveryPeriod represents the desired delivery time window.
// Matches proto: domain.delivery.common.v1.DeliveryPeriod
type DeliveryPeriod struct {
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

// OrderAssignedEvent represents an order assigned to a courier.
// Matches proto: domain.delivery.events.v1.PackageAssignedEvent
type OrderAssignedEvent struct {
	PackageID       string         `json:"package_id"`
	CourierID       string         `json:"courier_id"`
	Status          int32          `json:"status,omitempty"`
	AssignedAt      time.Time      `json:"assigned_at"`
	PickupAddress   Address        `json:"pickup_address"`
	DeliveryAddress Address        `json:"delivery_address"`
	DeliveryPeriod  DeliveryPeriod `json:"delivery_period,omitempty"`
	CustomerPhone   string         `json:"customer_phone,omitempty"`
	OccurredAt      time.Time      `json:"occurred_at,omitempty"`
}

// OrderAssignmentHandler handles order assignment events.
type OrderAssignmentHandler interface {
	HandleOrderAssigned(ctx context.Context, event OrderAssignedEvent) error
}

// DeliverySubscriberConfig holds configuration for the Kafka subscriber.
type DeliverySubscriberConfig struct {
	Brokers       []string
	ConsumerGroup string
}

// DefaultDeliverySubscriberConfig returns default configuration.
func DefaultDeliverySubscriberConfig() DeliverySubscriberConfig {
	return DeliverySubscriberConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: ConsumerGroupCourierEmulation,
	}
}

// DeliverySubscriber subscribes to delivery events from Kafka.
type DeliverySubscriber struct {
	subscriber message.Subscriber
	handler    OrderAssignmentHandler
	logger     watermill.LoggerAdapter
	stopCh     chan struct{}
}

// NewDeliverySubscriber creates a new Kafka delivery subscriber.
func NewDeliverySubscriber(
	config DeliverySubscriberConfig,
	handler OrderAssignmentHandler,
	logger watermill.LoggerAdapter,
) (*DeliverySubscriber, error) {
	if logger == nil {
		logger = watermill.NewStdLogger(false, false)
	}

	saramaConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               config.Brokers,
			Unmarshaler:           kafka.DefaultMarshaler{},
			ConsumerGroup:         config.ConsumerGroup,
			OverwriteSaramaConfig: saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("new kafka subscriber: %w", err)
	}

	return &DeliverySubscriber{
		subscriber: subscriber,
		handler:    handler,
		logger:     logger,
		stopCh:     make(chan struct{}),
	}, nil
}

// Start starts consuming messages from the order assigned topic.
func (s *DeliverySubscriber) Start(ctx context.Context) error {
	messages, err := s.subscriber.Subscribe(ctx, TopicOrderAssigned)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", TopicOrderAssigned, err)
	}

	go s.processMessages(ctx, messages)

	return nil
}

// processMessages processes incoming messages.
func (s *DeliverySubscriber) processMessages(ctx context.Context, messages <-chan *message.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case msg := <-messages:
			if msg == nil {
				continue
			}

			var event OrderAssignedEvent
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				s.logger.Error("Failed to unmarshal order assigned event", err, nil)
				msg.Nack()

				continue
			}

			err := s.handler.HandleOrderAssigned(ctx, event)
			if err != nil {
				s.logger.Error("Failed to handle order assigned event", err, nil)
				msg.Nack()

				continue
			}

			msg.Ack()
		}
	}
}

// Stop stops the subscriber.
func (s *DeliverySubscriber) Stop() error {
	close(s.stopCh)
	err := s.subscriber.Close()
	if err != nil {
		return fmt.Errorf("subscriber close: %w", err)
	}

	return nil
}

// DeliverySimulatorInterface defines the interface for starting deliveries.
type DeliverySimulatorInterface interface {
	StartDelivery(ctx context.Context, courierID string, order vo.DeliveryOrder) error
}

// CourierEmulationHandler implements OrderAssignmentHandler using DeliverySimulator.
type CourierEmulationHandler struct {
	deliverySimulator DeliverySimulatorInterface
}

// NewCourierEmulationHandler creates a new handler with the delivery simulator.
func NewCourierEmulationHandler(deliverySimulator DeliverySimulatorInterface) *CourierEmulationHandler {
	return &CourierEmulationHandler{
		deliverySimulator: deliverySimulator,
	}
}

// HandleOrderAssigned handles an order assignment by starting a delivery simulation.
func (h *CourierEmulationHandler) HandleOrderAssigned(ctx context.Context, event OrderAssignedEvent) error {
	// Extract coordinates from Address objects
	pickup, err := vo.NewLocation(event.PickupAddress.Latitude, event.PickupAddress.Longitude)
	if err != nil {
		return fmt.Errorf("pickup location: %w", err)
	}

	delivery, err := vo.NewLocation(event.DeliveryAddress.Latitude, event.DeliveryAddress.Longitude)
	if err != nil {
		return fmt.Errorf("delivery location: %w", err)
	}

	// PackageID is required in the new format
	order := vo.NewDeliveryOrder(
		event.PackageID, // Use PackageID as OrderID (they map 1:1 in this context)
		event.PackageID,
		pickup,
		delivery,
		event.AssignedAt,
	)

	if err := h.deliverySimulator.StartDelivery(ctx, event.CourierID, order); err != nil {
		return fmt.Errorf("start delivery: %w", err)
	}

	return nil
}
