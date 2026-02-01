package pkg_di

import (
	"time"

	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/config"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
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
	viper.SetDefault("SIMULATION_SPEED_KMH", 30.0)
	viper.SetDefault("SIMULATION_TIME_MULTIPLIER", 1.0)
	viper.SetDefault("SIMULATION_PICKUP_WAIT", 30*time.Second)
	viper.SetDefault("SIMULATION_DELIVERY_WAIT", 60*time.Second)
	viper.SetDefault("SIMULATION_FAILURE_RATE", 0.05)

	// Read configuration
	updateInterval := cfg.GetDuration("SIMULATION_UPDATE_INTERVAL")
	speedKmH := cfg.GetFloat64("SIMULATION_SPEED_KMH")
	timeMultiplier := cfg.GetFloat64("SIMULATION_TIME_MULTIPLIER")
	pickupWait := cfg.GetDuration("SIMULATION_PICKUP_WAIT")
	deliveryWait := cfg.GetDuration("SIMULATION_DELIVERY_WAIT")
	failureRate := cfg.GetFloat64("SIMULATION_FAILURE_RATE")

	config := services.DeliverySimulatorConfig{
		UpdateInterval:   updateInterval,
		SpeedKmH:         speedKmH,
		TimeMultiplier:   timeMultiplier,
		PickupWaitTime:   pickupWait,
		DeliveryWaitTime: deliveryWait,
		FailureRate:      failureRate,
	}

	return services.NewDeliverySimulator(config, routeGen, locationPub, statusPub)
}
