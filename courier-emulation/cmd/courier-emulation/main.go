/*
Shortlink application

Shop boundary
Courier Emulation service
*/
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/shortlink-org/go-sdk/graceful_shutdown"
	courier_di "github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/di"
	"github.com/spf13/viper"
)

// gracefulShutdownExitCode matches the conventional SIGTERM exit status (128 + 15).
const gracefulShutdownExitCode = 143

func main() {
	os.Exit(run())
}

func run() int {
	viper.SetDefault("SERVICE_NAME", "shortlink-courier-emulation")

	// Init a new service
	service, cleanup, err := courier_di.InitializeCourierEmulationService()
	if err != nil {
		panic(err)
	}

	service.Log.Info("Courier Emulation Service initialized")

	defer func() {
		if r := recover(); r != nil {
			service.Log.Error(fmt.Sprint(r))
		}
	}()

	// Create context for subscriber that can be canceled on shutdown
	ctx, cancel := context.WithCancelCause(context.Background())

	// Start the delivery subscriber to consume package assignment events.
	err = service.DeliverySubscriber.Start(ctx)
	if err != nil {
		service.Log.Error("Failed to start delivery subscriber", slog.String("error", err.Error()))
		cancel(fmt.Errorf("delivery subscriber start failed: %w", err)) //nolint:err113 // startup error should be attached to context cause
		cleanup()

		return 1
	}

	service.Log.Info("Delivery subscriber started, listening for package assignments")

	service.Log.Info("Courier Emulation Service running")

	// Handle SIGINT, SIGQUIT and SIGTERM - blocks until signal received
	signal := graceful_shutdown.GracefulShutdown()

	// Cancel the subscriber context to signal it to stop
	cancel(fmt.Errorf("shutdown signal received: %s", signal)) //nolint:err113 // dynamic message for shutdown reason

	// Run cleanup (stops simulations and closes publishers)
	cleanup()

	service.Log.Info("Courier Emulation Service stopped", slog.String("signal", signal.String()))

	// Exit Code 143: Graceful Termination (SIGTERM)

	return gracefulShutdownExitCode
}
