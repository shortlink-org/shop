package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

func TestRouteGenerator_GenerateRoute(t *testing.T) {
	// Create mock OSRM server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OSRMResponse{
			Code: "Ok",
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry string  `json:"geometry"`
			}{
				{
					Distance: 1885.4,
					Duration: 259.5,
					Geometry: "_p~iF~ps|U_ulLnnqC",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create generator with mock server
	config := RouteGeneratorConfig{
		OSRMBaseURL: server.URL,
		Timeout:     5 * time.Second,
	}
	generator := NewRouteGenerator(config)

	// Test route generation
	origin := vo.MustNewLocation(52.517037, 13.388860)
	destination := vo.MustNewLocation(52.529407, 13.397634)

	route, err := generator.GenerateRoute(context.Background(), origin, destination)

	require.NoError(t, err)
	assert.NotEmpty(t, route.ID())
	assert.Equal(t, origin, route.Origin())
	assert.Equal(t, destination, route.Destination())
	assert.InDelta(t, 1885.4, route.Distance(), 0.1)
	assert.Equal(t, 259*time.Second, route.Duration().Truncate(time.Second))
}

func TestRouteGenerator_GenerateRoute_NoRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OSRMResponse{
			Code:   "NoRoute",
			Routes: nil,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := RouteGeneratorConfig{
		OSRMBaseURL: server.URL,
		Timeout:     5 * time.Second,
	}
	generator := NewRouteGenerator(config)

	origin := vo.MustNewLocation(52.517037, 13.388860)
	destination := vo.MustNewLocation(52.529407, 13.397634)

	_, err := generator.GenerateRoute(context.Background(), origin, destination)
	assert.ErrorIs(t, err, ErrNoRouteFound)
}

func TestRouteGenerator_GenerateRoute_ServiceUnavailable(t *testing.T) {
	config := RouteGeneratorConfig{
		OSRMBaseURL: "http://localhost:59999", // Invalid port
		Timeout:     1 * time.Second,
	}
	generator := NewRouteGenerator(config)

	origin := vo.MustNewLocation(52.517037, 13.388860)
	destination := vo.MustNewLocation(52.529407, 13.397634)

	_, err := generator.GenerateRoute(context.Background(), origin, destination)
	assert.ErrorIs(t, err, ErrOSRMUnavailable)
}

func TestRouteGenerator_GenerateRandomRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OSRMResponse{
			Code: "Ok",
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry string  `json:"geometry"`
			}{
				{
					Distance: 2000.0,
					Duration: 300.0,
					Geometry: "_p~iF~ps|U_ulLnnqC",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := RouteGeneratorConfig{
		OSRMBaseURL: server.URL,
		Timeout:     5 * time.Second,
	}
	generator := NewRouteGenerator(config)

	bbox := vo.BerlinBoundingBox()
	route, err := generator.GenerateRandomRoute(context.Background(), bbox)

	require.NoError(t, err)
	assert.True(t, bbox.Contains(route.Origin()))
	assert.True(t, bbox.Contains(route.Destination()))
}

func TestRouteGenerator_GenerateBatch(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := OSRMResponse{
			Code: "Ok",
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry string  `json:"geometry"`
			}{
				{
					Distance: 2000.0,
					Duration: 300.0,
					Geometry: "_p~iF~ps|U_ulLnnqC",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := RouteGeneratorConfig{
		OSRMBaseURL: server.URL,
		Timeout:     5 * time.Second,
	}
	generator := NewRouteGenerator(config)

	bbox := vo.BerlinBoundingBox()
	routes, err := generator.GenerateBatch(context.Background(), bbox, 5)

	require.NoError(t, err)
	assert.Len(t, routes, 5)
	assert.Equal(t, 5, requestCount)
}

func TestDefaultRouteGeneratorConfig(t *testing.T) {
	config := DefaultRouteGeneratorConfig()

	assert.Equal(t, "http://localhost:5000", config.OSRMBaseURL)
	assert.Equal(t, 10*time.Second, config.Timeout)
}
