package oms_di

import (
	"context"
	"log/slog"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	"github.com/spf13/viper"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
	"github.com/shortlink-org/shop/oms/internal/usecases/order/event/on_delivery_status"
)

// NewDeliveryConsumer creates and starts a new Kafka delivery consumer for OMS.
func NewDeliveryConsumer(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	orderRepo ports.OrderRepository,
) (*kafka.DeliveryConsumer, func(), error) {
	viper.SetDefault("WATERMILL_KAFKA_BROKERS", []string{"localhost:9092"})

	brokers := cfg.GetStringSlice("WATERMILL_KAFKA_BROKERS")
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	consumerConfig := kafka.DeliveryConsumerConfig{
		Brokers: brokers,
		GroupID: kafka.ConsumerGroupOMSDelivery,
		Topic:   kafka.TopicDeliveryPackageStatus,
	}

	// Create event handler
	handler, err := on_delivery_status.NewHandler(log, orderRepo)
	if err != nil {
		return nil, func() {}, err
	}

	consumer, err := kafka.NewDeliveryConsumer(consumerConfig, handler, log)
	if err != nil {
		log.Warn("Failed to create Kafka delivery consumer, running without event consumption")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

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
