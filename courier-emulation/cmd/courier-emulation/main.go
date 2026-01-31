/*
Shortlink application

Shop boundary
Courier Emulation service
*/
package main

import (
	"log/slog"
	"os"

	"github.com/shortlink-org/go-sdk/graceful_shutdown"
	"github.com/spf13/viper"

	courier_di "github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/di"
)

func main() {
	viper.SetDefault("SERVICE_NAME", "shortlink-courier-emulation")

	// Init a new service
	service, cleanup, err := courier_di.InitializeCourierEmulationService()
	if err != nil {
		panic(err)
	}

	service.Log.Info("Courier Emulation Service initialized")

	defer func() {
		if r := recover(); r != nil {
			service.Log.Error(r.(string)) //nolint:forcetypeassert,errcheck // simple type assertion
		}
	}()

	// Handle SIGINT, SIGQUIT and SIGTERM.
	signal := graceful_shutdown.GracefulShutdown()

	cleanup()

	service.Log.Info("Courier Emulation Service stopped", slog.String("signal", signal.String()))

	// Exit Code 143: Graceful Termination (SIGTERM)
	os.Exit(143) //nolint:gocritic // exit code 143 is used to indicate graceful termination
}
