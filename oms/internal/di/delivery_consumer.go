package oms_di

import (
	"context"
	"log/slog"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	sdkkafka "github.com/shortlink-org/go-sdk/watermill/backends/kafka"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/event/on_delivery_status"
)

// NewDeliveryConsumer creates and starts a new Kafka delivery consumer for OMS.
func NewDeliveryConsumer(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
	publisher ports.EventPublisher,
) (*kafka.DeliveryConsumer, func(), error) {
	cfg.SetDefault("WATERMILL_KAFKA_CONSUMER_GROUP", kafka.ConsumerGroupOMSDelivery)

	// Create event handler
	handler, err := on_delivery_status.NewHandler(log, uow, orderRepo, publisher)
	if err != nil {
		return nil, func() {}, err
	}

	subscriber, err := sdkkafka.NewSubscriberFromConfig(log, cfg)
	if err != nil {
		log.Warn("Failed to create Kafka delivery subscriber, running without event consumption")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

	consumer := kafka.NewDeliveryConsumer(kafka.TopicDeliveryPackageStatus, subscriber, handler, log)

	// Start consuming in background
	if err := consumer.Start(ctx); err != nil {
		log.Warn("Failed to start Kafka delivery consumer", slog.Any("error", err))
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

	log.Info("Kafka delivery consumer started")

	cleanup := func() {
		if consumer != nil {
			err := consumer.Close()
			if err != nil {
				log.Warn("failed to close delivery consumer", slog.String("error", err.Error()))
			}
		}
	}

	return consumer, cleanup, nil
}
