package pkg_di

import (
	"time"

	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/config"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
)

// NewCourierSimulator creates the courier simulator.
func NewCourierSimulator(cfg *config.Config, routeGen *services.RouteGenerator, publisher *kafka.LocationPublisher) *services.CourierSimulator {
	viper.SetDefault("SIMULATION_UPDATE_INTERVAL", 5*time.Second)
	viper.SetDefault("SIMULATION_SPEED_KMH", 30.0)
	viper.SetDefault("SIMULATION_TIME_MULTIPLIER", 1.0)

	updateInterval := cfg.GetDuration("SIMULATION_UPDATE_INTERVAL")
	speedKmH := cfg.GetFloat64("SIMULATION_SPEED_KMH")
	timeMultiplier := cfg.GetFloat64("SIMULATION_TIME_MULTIPLIER")

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
