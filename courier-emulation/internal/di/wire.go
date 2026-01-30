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
	newRouteGenerator,
	newCourierSimulator,

	// Infrastructure
	newLocationPublisher,

	NewCourierEmulationService,
)

// newRouteGenerator creates the route generator service
func newRouteGenerator(cfg *config.Config) *services.RouteGenerator {
	osrmURL := cfg.GetString("OSRM_URL")
	if osrmURL == "" {
		osrmURL = "http://localhost:5000"
	}

	timeout := cfg.GetDuration("OSRM_TIMEOUT")
	if timeout == 0 {
		timeout = services.DefaultRouteGeneratorConfig().Timeout
	}

	return services.NewRouteGenerator(services.RouteGeneratorConfig{
		OSRMBaseURL: osrmURL,
		Timeout:     timeout,
	})
}

// newLocationPublisher creates the Kafka location publisher
func newLocationPublisher(cfg *config.Config, log logger.Logger) (*kafka.LocationPublisher, error) {
	brokers := cfg.GetStringSlice("KAFKA_BROKERS")
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	publisher, err := kafka.NewLocationPublisher(kafka.PublisherConfig{
		Brokers: brokers,
	}, nil)

	if err != nil {
		log.Warn("Failed to create Kafka publisher, running without Kafka")
		return nil, nil //nolint:nilerr // intentionally returning nil to continue without Kafka
	}

	return publisher, nil
}

// newCourierSimulator creates the courier simulator
func newCourierSimulator(cfg *config.Config, routeGen *services.RouteGenerator, publisher *kafka.LocationPublisher) *services.CourierSimulator {
	defaultCfg := services.DefaultCourierSimulatorConfig()

	updateInterval := cfg.GetDuration("SIMULATION_UPDATE_INTERVAL")
	if updateInterval == 0 {
		updateInterval = defaultCfg.UpdateInterval
	}

	speedKmH := cfg.GetFloat64("SIMULATION_SPEED_KMH")
	if speedKmH == 0 {
		speedKmH = defaultCfg.SpeedKmH
	}

	timeMultiplier := cfg.GetFloat64("SIMULATION_TIME_MULTIPLIER")
	if timeMultiplier == 0 {
		timeMultiplier = defaultCfg.TimeMultiplier
	}

	return services.NewCourierSimulator(
		services.CourierSimulatorConfig{
			UpdateInterval: updateInterval,
			SpeedKmH:       speedKmH,
			TimeMultiplier: timeMultiplier,
		},
		routeGen,
		publisher,
	)
}

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

		// Close Kafka publisher
		if publisher != nil {
			if err := publisher.Close(); err != nil {
				log.Error(err.Error())
			}
		}
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
