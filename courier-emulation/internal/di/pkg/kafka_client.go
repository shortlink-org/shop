package pkg_di

import (
	"log/slog"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	sdkkafka "github.com/shortlink-org/go-sdk/watermill/backends/kafka"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
	"github.com/spf13/viper"
)

// NewLocationPublisher creates the Kafka location publisher using go-sdk/watermill.
func NewLocationPublisher(cfg *config.Config, log logger.Logger) (*kafka.LocationPublisher, func(), error) {
	viper.SetDefault("WATERMILL_KAFKA_BROKERS", []string{"localhost:9092"})

	publisher, err := sdkkafka.NewPublisherFromConfig(log, cfg)
	if err != nil {
		log.Warn("Failed to create Kafka publisher, running without Kafka")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

	cleanup := func() {
		if publisher != nil {
			err := publisher.Close()
			if err != nil {
				log.Warn("failed to close location publisher", slog.String("error", err.Error()))
			}
		}
	}

	return kafka.NewLocationPublisher(publisher), cleanup, nil
}
