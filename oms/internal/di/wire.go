//go:generate wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
OMS DI-package
*/
package oms_di

import (
	"context"

	"github.com/authzed/authzed-go/v1"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"

	"github.com/shortlink-org/go-sdk/auth/permission"
	config "github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/flags"
	grpc "github.com/shortlink-org/go-sdk/grpc"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	profiling "github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"
	"github.com/shortlink-org/go-sdk/temporal"
	cartRPC "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1"
	orderRPC "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/run"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart"
	"github.com/shortlink-org/shop/oms/internal/usecases/order"
)

type OMSService struct {
	// Common
	Log    logger.Logger
	Config *config.Config

	// Observability
	Tracer        trace.TracerProvider
	Monitoring    *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Security
	authPermission *authzed.Client

	// Delivery
	run            *run.Response
	cartRPCServer  *cartRPC.CartRPC
	orderRPCServer *orderRPC.OrderRPC

	// Applications
	cartService *cart.UC

	// Temporal
	temporalClient client.Client
}

// OMSService ==========================================================================================================
// CustomDefaultSet - DefaultSet with go-sdk packages (config, context, flags, profiling)
var CustomDefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	newGoSDKProfiling,
	permission.New, // For authzed.Client
)

var OMSSet = wire.NewSet(
	// Common (custom DefaultSet)
	CustomDefaultSet,
	newGRPCServer,

	// Config & Observability (go-sdk)
	newGoSDKConfig,
	newGoSDKLogger,

	// Observability (go-sdk)
	newGoSDKTracer,
	newGoSDKMonitoring,

	// Delivery
	cartRPC.New,
	orderRPC.New,
	NewRunRPCServer,

	// Applications
	cart.New,
	order.New,

	// Temporal
	temporal.New,

	NewOMSService,
)

// newGRPCServer creates a gRPC server using go-sdk/grpc
func newGRPCServer(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, monitoring *metrics.Monitoring, cfg *config.Config) (*grpc.Server, error) {
	return grpc.InitServer(ctx, log, tracer, monitoring.Prometheus, nil, cfg)
}

// NewRunRPCServer starts the gRPC server
func NewRunRPCServer(runRPCServer *grpc.Server, _ *cartRPC.CartRPC) (*run.Response, error) {
	return run.Run(runRPCServer)
}

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

func NewOMSService(
	// Common
	log logger.Logger,
	config *config.Config,

	// Observability
	monitoring *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,

	// Security
	authPermission *authzed.Client,

	// Delivery
	run *run.Response,
	cartRPCServer *cartRPC.CartRPC,
	orderRPCServer *orderRPC.OrderRPC,

	// Temporal
	temporalClient client.Client,
) (*OMSService, error) {
	return &OMSService{
		// Common
		Log:    log,
		Config: config,

		// Observability
		Tracer:        tracer,
		Monitoring:    monitoring,
		PprofEndpoint: pprofHTTP,

		// Security
		// TODO: enable later
		// authPermission: authPermission,

		// Delivery
		run:            run,
		cartRPCServer:  cartRPCServer,
		orderRPCServer: orderRPCServer,

		// Temporal
		temporalClient: temporalClient,
	}, nil
}

func InitializeOMSService() (*OMSService, func(), error) {
	panic(wire.Build(OMSSet))
}
