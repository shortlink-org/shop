package pkg_di

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
)

// watermillLoggerAdapter adapts shortlink logger to Watermill logger interface.
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

// NewDeliverySubscriber creates the Kafka delivery subscriber with the handler.
func NewDeliverySubscriber(
	cfg *config.Config,
	log logger.Logger,
	deliverySimulator *services.DeliverySimulator,
) (*kafka.DeliverySubscriber, func(), error) {
	viper.SetDefault("WATERMILL_KAFKA_BROKERS", []string{"localhost:9092"})

	brokers := cfg.GetStringSlice("WATERMILL_KAFKA_BROKERS")
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	subscriberConfig := kafka.DeliverySubscriberConfig{
		Brokers:       brokers,
		ConsumerGroup: kafka.ConsumerGroupCourierEmulation,
	}

	// Create handler that connects to DeliverySimulator
	handler := kafka.NewCourierEmulationHandler(deliverySimulator)

	// Create Watermill logger adapter
	wmLogger := &watermillLoggerAdapter{log: log}

	subscriber, err := kafka.NewDeliverySubscriber(subscriberConfig, handler, wmLogger)
	if err != nil {
		log.Warn("Failed to create Kafka subscriber, running without event consumption")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

	cleanup := func() {
		if subscriber != nil {
			_ = subscriber.Stop()
		}
	}

	return subscriber, cleanup, nil
}
