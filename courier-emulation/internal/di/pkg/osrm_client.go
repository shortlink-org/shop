package pkg_di

import (
	"fmt"
	"time"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/spf13/viper"
)

// NewOSRMClient creates the OSRM route generator service.
func NewOSRMClient(cfg *config.Config) (*services.RouteGenerator, error) {
	viper.SetDefault("OSRM_URL", "http://localhost:5000")
	viper.SetDefault("OSRM_TIMEOUT", 10*time.Second)

	osrmURL := cfg.GetString("OSRM_URL")
	timeout := cfg.GetDuration("OSRM_TIMEOUT")

	rg, err := services.NewRouteGenerator(services.RouteGeneratorConfig{
		OSRMBaseURL: osrmURL,
		Timeout:     timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("new route generator: %w", err)
	}

	return rg, nil
}
