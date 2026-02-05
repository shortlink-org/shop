package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

const (
	// Cache configuration for route caching
	routeCacheNumCounters = 100_000   // track 100k routes
	routeCacheMaxCost     = 50_000_00 // ~50MB
	routeCacheBufferItems = 64
	routeCacheTTL         = 24 * time.Hour // routes rarely change
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
	cache      *ristretto.Cache[string, vo.Route]
}

// NewRouteGenerator creates a new RouteGenerator service.
func NewRouteGenerator(config RouteGeneratorConfig) (*RouteGenerator, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, vo.Route]{
		NumCounters: routeCacheNumCounters,
		MaxCost:     routeCacheMaxCost,
		BufferItems: routeCacheBufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create route cache: %w", err)
	}

	return &RouteGenerator{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		idCounter: 0,
		cache:     cache,
	}, nil
}

// Close closes the route generator and its cache.
func (rg *RouteGenerator) Close() {
	if rg.cache != nil {
		rg.cache.Close()
	}
}

// GenerateRoute generates a route between two locations using OSRM.
// Routes are cached by origin+destination coordinates for 24 hours.
func (rg *RouteGenerator) GenerateRoute(ctx context.Context, origin, destination vo.Location) (vo.Route, error) {
	// Create cache key from origin and destination coordinates
	cacheKey := fmt.Sprintf("%s:%s", origin.ToOSRMFormat(), destination.ToOSRMFormat())

	// Check cache first
	if cachedRoute, found := rg.cache.Get(cacheKey); found {
		return cachedRoute, nil
	}

	// Cache miss - fetch from OSRM
	route, err := rg.fetchRouteFromOSRM(ctx, origin, destination)
	if err != nil {
		return vo.Route{}, err
	}

	// Store in cache with TTL (cost=1 since all routes are similar size)
	rg.cache.SetWithTTL(cacheKey, route, 1, routeCacheTTL)

	return route, nil
}

// fetchRouteFromOSRM fetches a route from the OSRM API.
func (rg *RouteGenerator) fetchRouteFromOSRM(ctx context.Context, origin, destination vo.Location) (vo.Route, error) {
	url := fmt.Sprintf("%s/route/v1/driving/%s;%s?overview=full",
		rg.config.OSRMBaseURL,
		origin.ToOSRMFormat(),
		destination.ToOSRMFormat(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return vo.Route{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := rg.httpClient.Do(req)
	if err != nil {
		return vo.Route{}, fmt.Errorf("%w: %w", ErrOSRMUnavailable, err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			// Response body close error on defer, non-critical
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return vo.Route{}, fmt.Errorf("%w: status code %d", ErrOSRMUnavailable, resp.StatusCode)
	}

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return vo.Route{}, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return vo.Route{}, ErrNoRouteFound
	}

	osrmRoute := osrmResp.Routes[0]

	polyline, err := vo.NewPolyline(osrmRoute.Geometry)
	if err != nil {
		return vo.Route{}, fmt.Errorf("invalid polyline: %w", err)
	}

	rg.idCounter++
	routeID := fmt.Sprintf("route_%06d", rg.idCounter)

	route, err := vo.NewRoute(
		routeID,
		origin,
		destination,
		polyline,
		osrmRoute.Distance,
		time.Duration(osrmRoute.Duration)*time.Second,
	)
	if err != nil {
		return vo.Route{}, fmt.Errorf("new route: %w", err)
	}

	return route, nil
}

// GenerateRandomRoute generates a route between two random points in the bounding box.
func (rg *RouteGenerator) GenerateRandomRoute(ctx context.Context, bbox vo.BoundingBox) (vo.Route, error) {
	origin, destination := bbox.RandomPointPair()
	return rg.GenerateRoute(ctx, origin, destination)
}

// GenerateBatch generates multiple random routes in the bounding box.
func (rg *RouteGenerator) GenerateBatch(ctx context.Context, bbox vo.BoundingBox, count int) ([]vo.Route, error) {
	routes := make([]vo.Route, 0, count)

	for range count {
		select {
		case <-ctx.Done():
			return routes, fmt.Errorf("context: %w", ctx.Err())
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
