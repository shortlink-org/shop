package pkg_di

import (
	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	sdkkafka "github.com/shortlink-org/go-sdk/watermill/backends/kafka"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
)

// NewStatusPublisher creates the Kafka status publisher using go-sdk/watermill.
func NewStatusPublisher(cfg *config.Config, log logger.Logger) (*kafka.KafkaStatusPublisher, func(), error) {
	viper.SetDefault("WATERMILL_KAFKA_BROKERS", []string{"localhost:9092"})

	publisher, err := sdkkafka.NewPublisherFromConfig(log, cfg)
	if err != nil {
		log.Warn("Failed to create Kafka status publisher, running without Kafka")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

	cleanup := func() {
		if publisher != nil {
			_ = publisher.Close()
		}
	}

	return kafka.NewStatusPublisher(publisher), cleanup, nil
}
