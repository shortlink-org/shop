package pkg_di

import (
	"time"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
	"github.com/spf13/viper"
)

const (
	// defaultDeliverySimulationSpeedKmH is the baseline courier speed for delivery simulation.
	defaultDeliverySimulationSpeedKmH = 30.0
	// defaultDeliveryFailureRate is the default probability of a simulated failed delivery.
	defaultDeliveryFailureRate = 0.05
	// defaultPickupWait is the pause spent at pickup before switching to delivery.
	defaultPickupWait = 30 * time.Second
	// defaultDeliveryWait is the pause spent at the destination before completing delivery.
	defaultDeliveryWait = 60 * time.Second
)

// NewDeliverySimulator creates the delivery simulator with configuration.
func NewDeliverySimulator(
	cfg *config.Config,
	routeGen *services.RouteGenerator,
	locationPub *kafka.LocationPublisher,
	statusPub *kafka.KafkaStatusPublisher,
) *services.DeliverySimulator {
	// Set defaults
	viper.SetDefault("SIMULATION_UPDATE_INTERVAL", 5*time.Second)
	viper.SetDefault("SIMULATION_SPEED_KMH", defaultDeliverySimulationSpeedKmH)
	viper.SetDefault("SIMULATION_TIME_MULTIPLIER", 1.0)
	viper.SetDefault("SIMULATION_PICKUP_WAIT", defaultPickupWait)
	viper.SetDefault("SIMULATION_DELIVERY_WAIT", defaultDeliveryWait)
	viper.SetDefault("SIMULATION_FAILURE_RATE", defaultDeliveryFailureRate)

	// Read configuration
	updateInterval := cfg.GetDuration("SIMULATION_UPDATE_INTERVAL")
	speedKmH := cfg.GetFloat64("SIMULATION_SPEED_KMH")
	timeMultiplier := cfg.GetFloat64("SIMULATION_TIME_MULTIPLIER")
	pickupWait := cfg.GetDuration("SIMULATION_PICKUP_WAIT")
	deliveryWait := cfg.GetDuration("SIMULATION_DELIVERY_WAIT")
	failureRate := cfg.GetFloat64("SIMULATION_FAILURE_RATE")

	simCfg := services.DeliverySimulatorConfig{
		UpdateInterval:   updateInterval,
		SpeedKmH:         speedKmH,
		TimeMultiplier:   timeMultiplier,
		PickupWaitTime:   pickupWait,
		DeliveryWaitTime: deliveryWait,
		FailureRate:      failureRate,
	}

	return services.NewDeliverySimulator(simCfg, routeGen, locationPub, statusPub)
}
