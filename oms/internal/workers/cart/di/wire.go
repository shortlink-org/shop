//go:generate wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
OMS Cart Worker DI-package
*/
package oms_cart_worker_di

import (
	"context"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	config "github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/flags"
	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	profiling "github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"
	"github.com/shortlink-org/shop/oms/internal/loggeradapter"
	"github.com/shortlink-org/shop/oms/internal/workers/cart/cart_worker"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/temporal"
	shortlogger "github.com/shortlink-org/shortlink/pkg/logger"
	old_monitoring "github.com/shortlink-org/shortlink/pkg/observability/monitoring"
)

type OMSCartWorkerService struct {
	// Common
	Log    logger.Logger
	Config *config.Config

	// Observability
	Tracer        trace.TracerProvider
	Monitoring    *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Temporal
	temporalClient client.Client
	cartWorker     worker.Worker
}

// OMSCartWorkerService ================================================================================================
var CustomDefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	legacyLoggerAdapter,
)

var OMSCartWorkerSet = wire.NewSet(
	CustomDefaultSet,

	// Config & Observability (go-sdk)
	newGoSDKConfig,
	newGoSDKLogger,
	newGoSDKTracer,
	newGoSDKProfiling,
	newGoSDKMonitoring,
	legacyMonitoringFromGoSDK,

	// Temporal
	temporal.New,
	cart_worker.New,

	NewOMSCartWorkerService,
)

// newGoSDKConfig creates a go-sdk config instance
func newGoSDKConfig() (*config.Config, error) {
	return config.New()
}

func newGoSDKProfiling(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, cfg *config.Config) (profiling.PprofEndpoint, error) {
	return profiling.New(ctx, log, tracer, cfg)
}

func newGoSDKLogger(ctx context.Context, cfg *config.Config) (logger.Logger, func(), error) {
	return logger.NewDefault(ctx, cfg)
}

func newGoSDKTracer(ctx context.Context, log logger.Logger, cfg *config.Config) (trace.TracerProvider, func(), error) {
	return tracing.New(ctx, log, cfg)
}

func newGoSDKMonitoring(ctx context.Context, log logger.Logger, tracer trace.TracerProvider, cfg *config.Config) (*metrics.Monitoring, func(), error) {
	return metrics.New(ctx, log, tracer, cfg)
}

func NewOMSCartWorkerService(
	// Common
	log logger.Logger,
	config *config.Config,

	// Observability
	monitoring *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,

	// Temporal
	temporalClient client.Client,
	cartWorker worker.Worker,
) (*OMSCartWorkerService, error) {
	return &OMSCartWorkerService{
		// Common
		Log:    log,
		Config: config,

		// Observability
		Tracer:        tracer,
		Monitoring:    monitoring,
		PprofEndpoint: pprofHTTP,

		// Temporal
		temporalClient: temporalClient,
		cartWorker:     cartWorker,
	}, nil
}

func legacyLoggerAdapter(log logger.Logger) (shortlogger.Logger, func(), error) {
	return loggeradapter.New(log), func() {}, nil
}

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

func InitializeOMSCartWorkerService() (*OMSCartWorkerService, func(), error) {
	panic(wire.Build(OMSCartWorkerSet))
}
