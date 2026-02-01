package pkg_di

import (
	"time"

	"github.com/spf13/viper"

	"github.com/shortlink-org/go-sdk/config"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
)

// NewOSRMClient creates the OSRM route generator service.
func NewOSRMClient(cfg *config.Config) (*services.RouteGenerator, error) {
	viper.SetDefault("OSRM_URL", "http://localhost:5000")
	viper.SetDefault("OSRM_TIMEOUT", 10*time.Second)

	osrmURL := cfg.GetString("OSRM_URL")
	timeout := cfg.GetDuration("OSRM_TIMEOUT")

	return services.NewRouteGenerator(services.RouteGeneratorConfig{
		OSRMBaseURL: osrmURL,
		Timeout:     timeout,
	})
}
