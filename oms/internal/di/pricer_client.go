package oms_di

import (
	"log/slog"
	"time"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	"github.com/spf13/viper"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	pricer "github.com/shortlink-org/shop/oms/internal/infrastructure/grpc/pricer"
)

// NewPricerClient creates a new Pricer gRPC client.
func NewPricerClient(
	cfg *config.Config,
	log logger.Logger,
) (ports.PricerClient, func(), error) {
	viper.SetDefault("GRPC_CLIENT_TIMEOUT", "15s")

	// Use internal-gateway for all gRPC requests
	address := cfg.GetString("GRPC_CLIENT_HOST")
	if address == "" {
		address = "internal-gateway-istio.istio-ingress.svc.cluster.local"
	}

	timeout, err := time.ParseDuration(cfg.GetString("GRPC_CLIENT_TIMEOUT"))
	if err != nil {
		timeout = 15 * time.Second
	}

	clientCfg := pricer.Config{
		Address:    address,
		Timeout:    timeout,
		TLSEnabled: cfg.GetBool("GRPC_CLIENT_TLS_ENABLED"),
		CertPath:   cfg.GetString("GRPC_CLIENT_CERT_PATH"),
	}

	client, err := pricer.NewClient(clientCfg)
	if err != nil {
		log.Warn("Failed to create Pricer gRPC client, running without pricing integration")
		return nil, func() {}, nil //nolint:nilerr // intentionally returning nil to continue without Pricer
	}

	cleanup := func() {
		if client != nil {
			err := client.Close()
			if err != nil {
				log.Warn("failed to close pricer client", slog.String("error", err.Error()))
			}
		}
	}

	return client, cleanup, nil
}
