//go:generate wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
Pricer DI-package
*/
package di

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/flags"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	"github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"
	"github.com/shortlink-org/shop/pricer/internal/application"
	pkg_di "github.com/shortlink-org/shop/pricer/internal/di/pkg"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/cli"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/policy_evaluator"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/rpc/run"
	"github.com/shortlink-org/shop/pricer/internal/loggeradapter"
	shortlogger "github.com/shortlink-org/shortlink/pkg/logger"
	old_monitoring "github.com/shortlink-org/shortlink/pkg/observability/monitoring"
	"github.com/shortlink-org/shortlink/pkg/rpc"
)

type PricerService struct {
	// Common
	Log    logger.Logger
	Config *config.Config

	// Observability
	Tracer        trace.TracerProvider
	Monitoring    *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Delivery
	run *run.Response

	// Application
	CartService *application.CartService

	// CLI
	CLIHandler *cli.CLIHandler
}

// PricerService =======================================================================================================
// CustomDefaultSet - DefaultSet with go-sdk packages (config, context, flags, profiling)
var CustomDefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	legacyLoggerAdapter, // Required for shortlink legacy packages (rpc.InitServer, etc.)
	newGoSDKProfiling,
	// cache.New, - not used in pricer
	// permission.New, - not used in pricer
)

var PricerSet = wire.NewSet(
	// Common (custom DefaultSet with go-sdk packages)
	CustomDefaultSet,
	rpc.InitServer,
	pkg_di.ReadConfig,

	// Config & Observability (go-sdk)
	newGoSDKConfig,
	newGoSDKLogger,

	// Observability (go-sdk) - for PricerService
	newGoSDKTracer,
	newGoSDKMonitoring,
	legacyMonitoringFromGoSDK,

	// Repository
	newDiscountPolicy,
	newTaxPolicy,
	newPolicyNames,

	// Delivery
	NewRunRPCServer,

	// Application
	application.NewCartService,
	newCLIHandler,

	NewPricerService,
)

// newGoSDKConfig creates a go-sdk config instance
func newGoSDKConfig() (*config.Config, error) {
	return config.New()
}

// newGoSDKLogger creates a go-sdk logger instance for observability
func newGoSDKLogger(ctx context.Context, cfg *config.Config) (logger.Logger, func(), error) {
	return logger.NewDefault(ctx, cfg)
}

// newGoSDKTracer creates a tracer using go-sdk observability
func newGoSDKTracer(ctx context.Context, log logger.Logger, cfg *config.Config) (trace.TracerProvider, func(), error) {
	return tracing.New(ctx, log, cfg)
}

// newGoSDKMonitoring creates monitoring using go-sdk observability
func newGoSDKMonitoring(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, cfg *config.Config) (*metrics.Monitoring, func(), error) {
	return metrics.New(ctx, log, tracer, cfg)
}

// newGoSDKProfiling creates profiling endpoint using go-sdk observability
func newGoSDKProfiling(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, cfg *config.Config) (profiling.PprofEndpoint, error) {
	return profiling.New(ctx, log, tracer, cfg)
}

// legacyMonitoringFromGoSDK converts go-sdk monitoring to legacy shortlink monitoring structure.
func legacyMonitoringFromGoSDK(modern *metrics.Monitoring) *old_monitoring.Monitoring {
	if modern == nil {
		return nil
	}

	return &old_monitoring.Monitoring{
		Handler:    modern.Handler,
		Prometheus: modern.Prometheus,
		Metrics:    modern.Metrics,
	}
}

// TODO: refactoring. maybe drop this function
func NewRunRPCServer(runRPCServer *rpc.Server) (*run.Response, error) {
	return run.Run(runRPCServer)
}

// newDiscountPolicy creates a new DiscountPolicy
func newDiscountPolicy(ctx context.Context, log logger.Logger, cfg *pkg_di.Config) application.DiscountPolicy {
	discountPolicyPath := viper.GetString("policies.discounts")
	discountQuery := viper.GetString("queries.discounts")

	discountEvaluator, err := policy_evaluator.NewOPAEvaluator(log, discountPolicyPath, discountQuery)
	if err != nil {
		log.ErrorWithContext(ctx, "Failed to initialize discount policy evaluator", slog.Any("error", err))
	}

	return discountEvaluator
}

// newTaxPolicy creates a new TaxPolicy
func newTaxPolicy(ctx context.Context, log logger.Logger, cfg *pkg_di.Config) application.TaxPolicy {
	taxPolicyPath := viper.GetString("policies.taxes")
	taxQuery := viper.GetString("queries.taxes")

	taxEvaluator, err := policy_evaluator.NewOPAEvaluator(log, taxPolicyPath, taxQuery)
	if err != nil {
		log.ErrorWithContext(ctx, "Failed to initialize tax policy evaluator", slog.Any("error", err))
	}

	return taxEvaluator
}

// newPolicyNames retrieves policy names
func newPolicyNames(cfg *pkg_di.Config) ([]string, error) {
	discountPolicyPath := viper.GetString("policies.discounts")
	taxPolicyPath := viper.GetString("policies.taxes")

	return policy_evaluator.GetPolicyNames(discountPolicyPath, taxPolicyPath)
}

// newCLIHandler creates a new CLIHandler
func newCLIHandler(ctx context.Context, log logger.Logger, cartService *application.CartService, cfg *pkg_di.Config) *cli.CLIHandler {
	cartFiles := viper.GetStringSlice("cart_files")
	outputDir := viper.GetString("output_dir")

	discountParams := viper.GetStringMap("params.discount")
	taxParams := viper.GetStringMap("params.tax")

	cliHandler := &cli.CLIHandler{
		CartService: cartService,
		OutputDir:   outputDir,
	}

	// Process each cart file
	for _, cartFile := range cartFiles {
		fmt.Printf("Processing cart file: %s\n", cartFile)
		if err := cliHandler.Run(cartFile, discountParams, taxParams); err != nil {
			log.ErrorWithContext(ctx, "Error processing cart",
				slog.String("cart_file", cartFile),
				slog.Any("error", err),
			)
		}
	}

	return cliHandler
}

func legacyLoggerAdapter(log logger.Logger) (shortlogger.Logger, func(), error) {
	return loggeradapter.New(log), func() {}, nil
}

func NewPricerService(
	// Common
	log logger.Logger,
	config *config.Config,

	// Observability
	monitoring *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,

	// Delivery
	run *run.Response,

	// Application
	cartService *application.CartService,

	// CLI
	cliHandler *cli.CLIHandler,
) (*PricerService, error) {
	return &PricerService{
		// Common
		Log:    log,
		Config: config,

		// Observability
		Tracer:        tracer,
		Monitoring:    monitoring,
		PprofEndpoint: pprofHTTP,

		// Delivery
		run: run,

		// Application
		CartService: cartService,

		// CLI
		CLIHandler: cliHandler,
	}, nil
}

func InitializePricerService() (*PricerService, func(), error) {
	panic(wire.Build(PricerSet))
}
