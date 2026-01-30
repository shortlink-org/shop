package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

const (
	// TopicOrderAssigned is the Kafka topic for order assignment events from Delivery Service.
	TopicOrderAssigned = "delivery.order.assigned"
	// ConsumerGroupCourierEmulation is the consumer group for this service.
	ConsumerGroupCourierEmulation = "courier-emulation"
)

// OrderAssignedEvent represents an order assigned to a courier.
type OrderAssignedEvent struct {
	OrderID     string    `json:"order_id"`
	CourierID   string    `json:"courier_id"`
	PickupLat   float64   `json:"pickup_lat"`
	PickupLon   float64   `json:"pickup_lon"`
	DeliveryLat float64   `json:"delivery_lat"`
	DeliveryLon float64   `json:"delivery_lon"`
	AssignedAt  time.Time `json:"assigned_at"`
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
		return nil, err
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
		return err
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
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				s.logger.Error("Failed to unmarshal order assigned event", err, nil)
				msg.Nack()
				continue
			}

			if err := s.handler.HandleOrderAssigned(ctx, event); err != nil {
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
	return s.subscriber.Close()
}

// CourierEmulationHandler implements OrderAssignmentHandler.
type CourierEmulationHandler struct {
	startCourierFunc func(ctx context.Context, courierID string, pickup, delivery vo.Location) error
}

// NewCourierEmulationHandler creates a new handler.
func NewCourierEmulationHandler(
	startCourierFunc func(ctx context.Context, courierID string, pickup, delivery vo.Location) error,
) *CourierEmulationHandler {
	return &CourierEmulationHandler{
		startCourierFunc: startCourierFunc,
	}
}

// HandleOrderAssigned handles an order assignment by starting a courier simulation.
func (h *CourierEmulationHandler) HandleOrderAssigned(ctx context.Context, event OrderAssignedEvent) error {
	pickup, err := vo.NewLocation(event.PickupLat, event.PickupLon)
	if err != nil {
		return err
	}

	delivery, err := vo.NewLocation(event.DeliveryLat, event.DeliveryLon)
	if err != nil {
		return err
	}

	return h.startCourierFunc(ctx, event.CourierID, pickup, delivery)
}
