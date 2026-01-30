package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

// RouteGenerator errors
var (
	ErrOSRMUnavailable = errors.New("OSRM service unavailable")
	ErrNoRouteFound    = errors.New("no route found between points")
	ErrInvalidResponse = errors.New("invalid OSRM response")
)

// OSRMResponse represents the response from OSRM routing API.
type OSRMResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"` // meters
		Duration float64 `json:"duration"` // seconds
		Geometry string  `json:"geometry"` // encoded polyline
	} `json:"routes"`
}

// RouteGeneratorConfig holds configuration for the route generator.
type RouteGeneratorConfig struct {
	OSRMBaseURL string
	Timeout     time.Duration
}

// DefaultRouteGeneratorConfig returns default configuration.
func DefaultRouteGeneratorConfig() RouteGeneratorConfig {
	return RouteGeneratorConfig{
		OSRMBaseURL: "http://localhost:5000",
		Timeout:     10 * time.Second,
	}
}

// RouteGenerator is a domain service for generating routes via OSRM.
type RouteGenerator struct {
	config     RouteGeneratorConfig
	httpClient *http.Client
	idCounter  int
}

// NewRouteGenerator creates a new RouteGenerator service.
func NewRouteGenerator(config RouteGeneratorConfig) *RouteGenerator {
	return &RouteGenerator{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		idCounter: 0,
	}
}

// GenerateRoute generates a route between two locations using OSRM.
func (rg *RouteGenerator) GenerateRoute(ctx context.Context, origin, destination vo.Location) (vo.Route, error) {
	url := fmt.Sprintf("%s/route/v1/driving/%s;%s?overview=full",
		rg.config.OSRMBaseURL,
		origin.ToOSRMFormat(),
		destination.ToOSRMFormat(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return vo.Route{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := rg.httpClient.Do(req)
	if err != nil {
		return vo.Route{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return vo.Route{}, fmt.Errorf("%w: status code %d", ErrOSRMUnavailable, resp.StatusCode)
	}

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return vo.Route{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return vo.Route{}, ErrNoRouteFound
	}

	route := osrmResp.Routes[0]

	polyline, err := vo.NewPolyline(route.Geometry)
	if err != nil {
		return vo.Route{}, fmt.Errorf("invalid polyline: %w", err)
	}

	rg.idCounter++
	routeID := fmt.Sprintf("route_%06d", rg.idCounter)

	return vo.NewRoute(
		routeID,
		origin,
		destination,
		polyline,
		route.Distance,
		time.Duration(route.Duration)*time.Second,
	)
}

// GenerateRandomRoute generates a route between two random points in the bounding box.
func (rg *RouteGenerator) GenerateRandomRoute(ctx context.Context, bbox vo.BoundingBox) (vo.Route, error) {
	origin, destination := bbox.RandomPointPair()
	return rg.GenerateRoute(ctx, origin, destination)
}

// GenerateBatch generates multiple random routes in the bounding box.
func (rg *RouteGenerator) GenerateBatch(ctx context.Context, bbox vo.BoundingBox, count int) ([]vo.Route, error) {
	routes := make([]vo.Route, 0, count)

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return routes, ctx.Err()
		default:
		}

		route, err := rg.GenerateRandomRoute(ctx, bbox)
		if err != nil {
			// Skip failed routes, continue generating
			continue
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// HealthCheck checks if OSRM service is available.
func (rg *RouteGenerator) HealthCheck(ctx context.Context) error {
	// Test with a simple route in Berlin
	origin := vo.MustNewLocation(52.5200, 13.4050)
	destination := vo.MustNewLocation(52.5300, 13.4150)

	_, err := rg.GenerateRoute(ctx, origin, destination)
	return err
}
