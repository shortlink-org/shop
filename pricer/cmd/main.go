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
	viper.SetDefault("GRPC_SERVER_ENABLED", true)

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

	// CLI mode: when gRPC is disabled, process cart files
	if !service.Config.GetBool("GRPC_SERVER_ENABLED") {
		cartFiles := viper.GetStringSlice("cart_files")
		discountParams := viper.GetStringMap("params.discount")
		taxParams := viper.GetStringMap("params.tax")
		if discountParams == nil {
			discountParams = make(map[string]interface{})
		}
		if taxParams == nil {
			taxParams = make(map[string]interface{})
		}
		for _, cartFile := range cartFiles {
			if err := service.CLIHandler.Run(cartFile, discountParams, taxParams); err != nil {
				service.Log.Error("CLI processing failed", slog.String("cart_file", cartFile), slog.Any("error", err))
			}
		}
	}

	// Handle SIGINT, SIGQUIT and SIGTERM.
	signal := graceful_shutdown.GracefulShutdown()

	cleanup()

	service.Log.Info("Service stopped", slog.String("signal", signal.String()))

	// Exit Code 143: Graceful Termination (SIGTERM)
	os.Exit(143) //nolint:gocritic // exit code 143 is used to indicate graceful termination
}
