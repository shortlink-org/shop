package main

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/graceful_shutdown"
	"github.com/shortlink-org/shop/pricer/internal/di"
)

func main() {
	viper.SetDefault("SERVICE_NAME", "shop-pricer")

	// Init a new service
	service, cleanup, err := di.InitializePricerService()
	if err != nil {
		panic(err)
	}
	service.Log.Info("Service initialized")

	defer func() {
		if r := recover(); r != nil {
			service.Log.Error("panic recovered", slog.Any("error", r))
		}
	}()

	// Handle SIGINT, SIGQUIT and SIGTERM.
	signal := graceful_shutdown.GracefulShutdown()

	cleanup()

	service.Log.Info("Service stopped", slog.String("signal", signal.String()))

	// Exit Code 143: Graceful Termination (SIGTERM)
	os.Exit(143) //nolint:gocritic // exit code 143 is used to indicate graceful termination
}
