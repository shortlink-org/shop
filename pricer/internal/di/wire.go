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

	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/flags"
	"github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	"github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"
	pkg_di "github.com/shortlink-org/shop/pricer/internal/di/pkg"
	"github.com/shortlink-org/shop/pricer/internal/domain/pricing"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/cli"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/policy_evaluator"
	cartv1 "github.com/shortlink-org/shop/pricer/internal/infrastructure/rpc/cart/v1"
	"github.com/shortlink-org/shop/pricer/internal/infrastructure/rpc/run"
	"github.com/shortlink-org/shop/pricer/internal/usecases/cart/command/calculate_total"
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

	// Use cases
	CalculateTotalHandler *calculate_total.Handler

	// CLI
	CLIHandler *cli.CLIHandler
}

// PricerService =======================================================================================================
// CustomDefaultSet - DefaultSet with go-sdk packages (config, context, flags, profiling)
var CustomDefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	newGoSDKProfiling,
	// cache.New, - not used in pricer
	// permission.New, - not used in pricer
)

var PricerSet = wire.NewSet(
	// Common (custom DefaultSet with go-sdk packages)
	CustomDefaultSet,
	pkg_di.ReadConfig,

	// Config & Observability (go-sdk)
	newGoSDKConfig,
	newGoSDKLogger,

	// Observability (go-sdk) - for PricerService
	newGoSDKTracer,
	newGoSDKMonitoring,

	// gRPC Server (go-sdk)
	newGRPCServerWithHandler,

	// Repository
	newDiscountPolicy,
	newTaxPolicy,
	newPolicyNames,

	// Delivery
	NewRunRPCServer,

	// Use cases
	calculate_total.NewHandler,
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

// newGRPCServerWithHandler creates gRPC server and registers CartService handler
func newGRPCServerWithHandler(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, monitoring *metrics.Monitoring, cfg *config.Config, calculateTotalHandler *calculate_total.Handler) (*grpc.Server, error) {
	promRegistry := monitoring.Prometheus
	server, err := grpc.InitServer(ctx, log, tracer, promRegistry, nil, cfg)
	if err != nil {
		return nil, err
	}
	if server != nil {
		handler := cartv1.NewCartHandler(calculateTotalHandler)
		cartv1.RegisterCartServiceServer(server.Server, handler)
	}
	return server, nil
}

func NewRunRPCServer(runRPCServer *grpc.Server) (*run.Response, error) {
	return run.Run(runRPCServer)
}

// newDiscountPolicy creates a new discount policy
func newDiscountPolicy(ctx context.Context, log logger.Logger, cfg *pkg_di.Config) (*pricing.DiscountPolicy, error) {
	discountPolicyPath := viper.GetString("policies.discounts")
	discountQuery := viper.GetString("queries.discounts")

	evaluator, err := policy_evaluator.NewOPAEvaluator(log, discountPolicyPath, discountQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discount policy evaluator: %w", err)
	}

	return &pricing.DiscountPolicy{Evaluator: evaluator}, nil
}

// newTaxPolicy creates a new tax policy
func newTaxPolicy(ctx context.Context, log logger.Logger, cfg *pkg_di.Config) (*pricing.TaxPolicy, error) {
	taxPolicyPath := viper.GetString("policies.taxes")
	taxQuery := viper.GetString("queries.taxes")

	evaluator, err := policy_evaluator.NewOPAEvaluator(log, taxPolicyPath, taxQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tax policy evaluator: %w", err)
	}

	return &pricing.TaxPolicy{Evaluator: evaluator}, nil
}

// newPolicyNames retrieves policy names
func newPolicyNames(cfg *pkg_di.Config) ([]string, error) {
	discountPolicyPath := viper.GetString("policies.discounts")
	taxPolicyPath := viper.GetString("policies.taxes")

	return policy_evaluator.GetPolicyNames(discountPolicyPath, taxPolicyPath)
}

// newCLIHandler creates a new CLIHandler (does not run processing - use Run() explicitly for CLI mode)
func newCLIHandler(calculateTotalHandler *calculate_total.Handler, cfg *pkg_di.Config) *cli.CLIHandler {
	outputDir := viper.GetString("output_dir")
	return cli.NewCLIHandler(calculateTotalHandler, outputDir)
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

	// Use cases
	calculateTotalHandler *calculate_total.Handler,

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

		// Use cases
		CalculateTotalHandler: calculateTotalHandler,

		// CLI
		CLIHandler: cliHandler,
	}, nil
}

func InitializePricerService() (*PricerService, func(), error) {
	panic(wire.Build(PricerSet))
}
