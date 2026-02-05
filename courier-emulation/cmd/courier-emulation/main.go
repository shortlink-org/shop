/*
Shortlink application

Shop boundary
Courier Emulation service
*/
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/shortlink-org/go-sdk/graceful_shutdown"
	courier_di "github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/di"
	"github.com/spf13/viper"
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

	// Create context for subscriber that can be canceled on shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Start the delivery subscriber to consume order assignment events
	if service.DeliverySubscriber != nil {
		err := service.DeliverySubscriber.Start(ctx)
		if err != nil {
			service.Log.Error("Failed to start delivery subscriber", slog.String("error", err.Error()))
		} else {
			service.Log.Info("Delivery subscriber started, listening for order assignments")
		}
	} else {
		service.Log.Warn("Delivery subscriber not available, running without event consumption")
	}

	service.Log.Info("Courier Emulation Service running")

	// Handle SIGINT, SIGQUIT and SIGTERM - blocks until signal received
	signal := graceful_shutdown.GracefulShutdown()

	// Cancel the subscriber context to signal it to stop
	cancel()

	// Run cleanup (stops simulations and closes publishers)
	cleanup()

	service.Log.Info("Courier Emulation Service stopped", slog.String("signal", signal.String()))

	// Exit Code 143: Graceful Termination (SIGTERM)
	os.Exit(143) //nolint:gocritic // exit code 143 is used to indicate graceful termination
}
