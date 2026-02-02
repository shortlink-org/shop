package oms_di

import (
	"time"

	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	delivery "github.com/shortlink-org/shop/oms/internal/infrastructure/grpc/delivery"
)

// NewDeliveryClient creates a new Delivery gRPC client.
func NewDeliveryClient(
	cfg *config.Config,
	log logger.Logger,
) (ports.DeliveryClient, func(), error) {
	viper.SetDefault("DELIVERY_GRPC_ADDRESS", "delivery:50051")
	viper.SetDefault("DELIVERY_GRPC_TIMEOUT", "10s")

	address := cfg.GetString("DELIVERY_GRPC_ADDRESS")
	if address == "" {
		address = "delivery:50051"
	}

	timeout, err := time.ParseDuration(cfg.GetString("DELIVERY_GRPC_TIMEOUT"))
	if err != nil {
		timeout = 10 * time.Second
	}

	clientCfg := delivery.Config{
		Address: address,
		Timeout: timeout,
	}

	client, err := delivery.NewClient(clientCfg)
	if err != nil {
		log.Warn("Failed to create Delivery gRPC client, running without delivery integration")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Delivery
	}

	cleanup := func() {
		if client != nil {
			_ = client.Close()
		}
	}

	return client, cleanup, nil
}
