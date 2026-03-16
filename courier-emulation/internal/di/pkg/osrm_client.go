package pkg_di

import (
	"errors"
	"fmt"
	"time"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/services"
	"github.com/spf13/viper"
)

const defaultOSRMTimeout = 10 * time.Second

var errIncompleteOSRMAuthHeader = errors.New(
	"OSRM_AUTH_HEADER_NAME and OSRM_AUTH_HEADER_VALUE must be set together",
)

// NewOSRMClient creates the OSRM route generator service.
func NewOSRMClient(cfg *config.Config) (*services.RouteGenerator, error) {
	viper.SetDefault("OSRM_URL", "http://localhost:5000")
	viper.SetDefault("OSRM_TIMEOUT", defaultOSRMTimeout)

	osrmURL := cfg.GetString("OSRM_URL")
	timeout := cfg.GetDuration("OSRM_TIMEOUT")
	authHeaderName := cfg.GetString("OSRM_AUTH_HEADER_NAME")
	authHeaderValue := cfg.GetString("OSRM_AUTH_HEADER_VALUE")

	if (authHeaderName == "") != (authHeaderValue == "") {
		return nil, errIncompleteOSRMAuthHeader
	}

	routeGenerator, err := services.NewRouteGenerator(services.RouteGeneratorConfig{
		OSRMBaseURL:     osrmURL,
		Timeout:         timeout,
		AuthHeaderName:  authHeaderName,
		AuthHeaderValue: authHeaderValue,
	})
	if err != nil {
		return nil, fmt.Errorf("new route generator: %w", err)
	}

	return routeGenerator, nil
}
