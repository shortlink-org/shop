//go:generate go tool wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

/*
Courier Emulation DI-package
*/
package courier_di

import (
	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/go-sdk/config"
	sdkctx "github.com/shortlink-org/go-sdk/context"
	"github.com/shortlink-org/go-sdk/flags"
	"github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/go-sdk/observability/metrics"
	"github.com/shortlink-org/go-sdk/observability/profiling"
	"github.com/shortlink-org/go-sdk/observability/tracing"

	pkg_di "github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/di/pkg"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
)

type CourierEmulationService struct {
	// Common
	Log    logger.Logger
	Config *config.Config

	// Observability
	Tracer        trace.TracerProvider
	Monitoring    *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Domain services
	RouteGenerator   *services.RouteGenerator
	CourierSimulator *services.CourierSimulator

	// Infrastructure
	LocationPublisher *kafka.LocationPublisher
}

// DefaultSet ==========================================================================================================
var DefaultSet = wire.NewSet(
	sdkctx.New,
	flags.New,
	config.New,
	logger.NewDefault,
	tracing.New,
	metrics.New,
	profiling.New,
)

// CourierEmulationSet =================================================================================================
var CourierEmulationSet = wire.NewSet(
	// Common
	DefaultSet,

	// Domain services
	pkg_di.NewOSRMClient,
	pkg_di.NewCourierSimulator,

	// Infrastructure
	pkg_di.NewLocationPublisher,

	NewCourierEmulationService,
)

func NewCourierEmulationService(
	// Common
	log logger.Logger,
	cfg *config.Config,

	// Observability
	monitoring *metrics.Monitoring,
	tracer trace.TracerProvider,
	pprofHTTP profiling.PprofEndpoint,

	// Domain services
	routeGen *services.RouteGenerator,
	simulator *services.CourierSimulator,

	// Infrastructure
	publisher *kafka.LocationPublisher,
) (*CourierEmulationService, func(), error) {
	cleanup := func() {
		log.Info("Shutting down courier simulation...")

		// Stop all couriers
		simulator.Stop()

		// Note: Kafka publisher cleanup is handled by Wire via the cleanup function
		// returned by pkg_di.NewLocationPublisher
	}

	return &CourierEmulationService{
		// Common
		Log:    log,
		Config: cfg,

		// Observability
		Tracer:        tracer,
		Monitoring:    monitoring,
		PprofEndpoint: pprofHTTP,

		// Domain services
		RouteGenerator:   routeGen,
		CourierSimulator: simulator,

		// Infrastructure
		LocationPublisher: publisher,
	}, cleanup, nil
}

func InitializeCourierEmulationService() (*CourierEmulationService, func(), error) {
	panic(wire.Build(CourierEmulationSet))
}
