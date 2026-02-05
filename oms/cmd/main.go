/*
Shortlink application

boundary: OMS
service: shortlink-shop-oms
*/
package main

import (
	"log/slog"
	"os"

	"github.com/shortlink-org/go-sdk/graceful_shutdown"
	"github.com/spf13/viper"

	oms_di "github.com/shortlink-org/shop/oms/internal/di"
)

func main() {
	viper.SetDefault("SERVICE_NAME", "shortlink-shop-oms")

	// Init a new service
	service, cleanup, err := oms_di.InitializeOMSService()
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
