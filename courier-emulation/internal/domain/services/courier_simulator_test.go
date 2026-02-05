package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLocationPublisher is a mock implementation for testing.
type mockLocationPublisher struct {
	events []vo.CourierLocationEvent
	mu     sync.Mutex
	closed bool
}

func newMockLocationPublisher() *mockLocationPublisher {
	return &mockLocationPublisher{
		events: make([]vo.CourierLocationEvent, 0),
	}
}

func (m *mockLocationPublisher) PublishLocation(ctx context.Context, event vo.CourierLocationEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, event)

	return nil
}

func (m *mockLocationPublisher) Close() error {
	m.closed = true
	return nil
}

func (m *mockLocationPublisher) GetEvents() []vo.CourierLocationEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]vo.CourierLocationEvent, len(m.events))
	copy(result, m.events)

	return result
}

func TestCourierSimulator_StartCourierWithRoute(t *testing.T) {
	publisher := newMockLocationPublisher()
	config := CourierSimulatorConfig{
		UpdateInterval: 100 * time.Millisecond,
		SpeedKmH:       30.0,
		TimeMultiplier: 10.0, // 10x speed for faster testing
	}

	simulator := NewCourierSimulator(config, nil, publisher)
	defer simulator.Stop()

	// Create a simple route
	origin := vo.MustNewLocation(52.5200, 13.4050)
	destination := vo.MustNewLocation(52.5210, 13.4060)
	polyline := vo.MustNewPolyline("_c`|IgpvpAaB{A") // Simple encoded polyline

	route, err := vo.NewRoute("test-route", origin, destination, polyline, 150, 30*time.Second)
	require.NoError(t, err)

	// Start courier
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = simulator.StartCourierWithRoute(ctx, "courier-1", route)
	require.NoError(t, err)

	// Wait for some updates
	time.Sleep(500 * time.Millisecond)

	// Check courier state
	state, exists := simulator.GetCourierState("courier-1")
	require.True(t, exists)
	assert.Equal(t, "courier-1", state.ID)
	assert.Equal(t, "test-route", state.CurrentRoute.ID())
}

func TestCourierSimulator_GetAllCouriers(t *testing.T) {
	publisher := newMockLocationPublisher()
	config := DefaultCourierSimulatorConfig()
	config.UpdateInterval = 1 * time.Hour // Don't actually run updates

	simulator := NewCourierSimulator(config, nil, publisher)
	defer simulator.Stop()

	// Create routes for multiple couriers
	origin := vo.MustNewLocation(52.5200, 13.4050)
	destination := vo.MustNewLocation(52.5210, 13.4060)
	polyline := vo.MustNewPolyline("_c`|IgpvpAaB{A")

	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		route, _ := vo.NewRoute("route", origin, destination, polyline, 150, 30*time.Second)
		err := simulator.StartCourierWithRoute(ctx, "courier-"+string(rune('0'+i)), route)
		require.NoError(t, err)
	}

	couriers := simulator.GetAllCouriers()
	assert.Len(t, couriers, 3)
}

func TestCourierSimulator_StopCourier(t *testing.T) {
	publisher := newMockLocationPublisher()
	config := DefaultCourierSimulatorConfig()
	config.UpdateInterval = 1 * time.Hour

	simulator := NewCourierSimulator(config, nil, publisher)
	defer simulator.Stop()

	origin := vo.MustNewLocation(52.5200, 13.4050)
	destination := vo.MustNewLocation(52.5210, 13.4060)
	polyline := vo.MustNewPolyline("_c`|IgpvpAaB{A")
	route, _ := vo.NewRoute("route", origin, destination, polyline, 150, 30*time.Second)

	ctx := context.Background()
	_ = simulator.StartCourierWithRoute(ctx, "courier-1", route)

	// Verify courier exists
	_, exists := simulator.GetCourierState("courier-1")
	assert.True(t, exists)

	// Stop courier
	simulator.StopCourier("courier-1")

	// Verify courier is removed
	_, exists = simulator.GetCourierState("courier-1")
	assert.False(t, exists)
}

func TestInterpolateLocation(t *testing.T) {
	from := vo.MustNewLocation(52.5200, 13.4050)
	to := vo.MustNewLocation(52.5300, 13.4150)

	// Test midpoint
	mid := interpolateLocation(from, to, 0.5)
	assert.InDelta(t, 52.5250, mid.Latitude(), 0.0001)
	assert.InDelta(t, 13.4100, mid.Longitude(), 0.0001)

	// Test start point
	start := interpolateLocation(from, to, 0.0)
	assert.Equal(t, from.Latitude(), start.Latitude())
	assert.Equal(t, from.Longitude(), start.Longitude())

	// Test end point
	end := interpolateLocation(from, to, 1.0)
	assert.Equal(t, to.Latitude(), end.Latitude())
	assert.Equal(t, to.Longitude(), end.Longitude())
}

func TestCalculateHeading(t *testing.T) {
	tests := []struct {
		name     string
		from     vo.Location
		to       vo.Location
		expected float64
		delta    float64
	}{
		{
			name:     "North",
			from:     vo.MustNewLocation(52.5200, 13.4050),
			to:       vo.MustNewLocation(52.5300, 13.4050),
			expected: 0.0,
			delta:    1.0,
		},
		{
			name:     "East",
			from:     vo.MustNewLocation(52.5200, 13.4050),
			to:       vo.MustNewLocation(52.5200, 13.4150),
			expected: 90.0,
			delta:    1.0,
		},
		{
			name:     "South",
			from:     vo.MustNewLocation(52.5200, 13.4050),
			to:       vo.MustNewLocation(52.5100, 13.4050),
			expected: 180.0,
			delta:    1.0,
		},
		{
			name:     "West",
			from:     vo.MustNewLocation(52.5200, 13.4050),
			to:       vo.MustNewLocation(52.5200, 13.3950),
			expected: 270.0,
			delta:    1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heading := calculateHeading(tt.from, tt.to)
			assert.InDelta(t, tt.expected, heading, tt.delta)
		})
	}
}

func TestDefaultCourierSimulatorConfig(t *testing.T) {
	config := DefaultCourierSimulatorConfig()

	assert.Equal(t, 5*time.Second, config.UpdateInterval)
	assert.Equal(t, 30.0, config.SpeedKmH)
	assert.Equal(t, 1.0, config.TimeMultiplier)
}
