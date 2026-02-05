package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStatusPublisher is a mock implementation of StatusPublisher.
type mockStatusPublisher struct {
	mu             sync.Mutex
	pickupEvents   []kafka.PickUpOrderEvent
	deliveryEvents []kafka.DeliverOrderEvent
}

func newMockStatusPublisher() *mockStatusPublisher {
	return &mockStatusPublisher{
		pickupEvents:   make([]kafka.PickUpOrderEvent, 0),
		deliveryEvents: make([]kafka.DeliverOrderEvent, 0),
	}
}

func (m *mockStatusPublisher) PublishPickUp(ctx context.Context, event kafka.PickUpOrderEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pickupEvents = append(m.pickupEvents, event)

	return nil
}

func (m *mockStatusPublisher) PublishDelivery(ctx context.Context, event kafka.DeliverOrderEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deliveryEvents = append(m.deliveryEvents, event)

	return nil
}

func (m *mockStatusPublisher) Close() error {
	return nil
}

func (m *mockStatusPublisher) GetPickupEvents() []kafka.PickUpOrderEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]kafka.PickUpOrderEvent, len(m.pickupEvents))
	copy(result, m.pickupEvents)

	return result
}

func (m *mockStatusPublisher) GetDeliveryEvents() []kafka.DeliverOrderEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]kafka.DeliverOrderEvent, len(m.deliveryEvents))
	copy(result, m.deliveryEvents)

	return result
}

func TestDeliveryPhase_ToCourierStatus(t *testing.T) {
	tests := []struct {
		phase    vo.DeliveryPhase
		expected string
	}{
		{vo.PhaseIdle, vo.CourierStatusIdle},
		{vo.PhaseHeadingToPickup, vo.CourierStatusMoving},
		{vo.PhasePickingUp, vo.CourierStatusPickingUp},
		{vo.PhaseHeadingToCustomer, vo.CourierStatusMoving},
		{vo.PhaseDelivering, vo.CourierStatusDelivering},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.ToCourierStatus())
		})
	}
}

func TestDeliveryPhase_IsActive(t *testing.T) {
	assert.False(t, vo.PhaseIdle.IsActive())
	assert.True(t, vo.PhaseHeadingToPickup.IsActive())
	assert.True(t, vo.PhasePickingUp.IsActive())
	assert.True(t, vo.PhaseHeadingToCustomer.IsActive())
	assert.True(t, vo.PhaseDelivering.IsActive())
}

func TestDeliveryPhase_IsMoving(t *testing.T) {
	assert.False(t, vo.PhaseIdle.IsMoving())
	assert.True(t, vo.PhaseHeadingToPickup.IsMoving())
	assert.False(t, vo.PhasePickingUp.IsMoving())
	assert.True(t, vo.PhaseHeadingToCustomer.IsMoving())
	assert.False(t, vo.PhaseDelivering.IsMoving())
}

func TestDeliveryPhase_IsWaiting(t *testing.T) {
	assert.False(t, vo.PhaseIdle.IsWaiting())
	assert.False(t, vo.PhaseHeadingToPickup.IsWaiting())
	assert.True(t, vo.PhasePickingUp.IsWaiting())
	assert.False(t, vo.PhaseHeadingToCustomer.IsWaiting())
	assert.True(t, vo.PhaseDelivering.IsWaiting())
}

func TestDeliveryOrder_Creation(t *testing.T) {
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	assignedAt := time.Now()

	order := vo.NewDeliveryOrder("order-123", "pkg-456", pickup, delivery, assignedAt)

	assert.Equal(t, "order-123", order.OrderID())
	assert.Equal(t, "pkg-456", order.PackageID())
	assert.Equal(t, pickup.Latitude(), order.PickupLocation().Latitude())
	assert.Equal(t, pickup.Longitude(), order.PickupLocation().Longitude())
	assert.Equal(t, delivery.Latitude(), order.DeliveryLocation().Latitude())
	assert.Equal(t, delivery.Longitude(), order.DeliveryLocation().Longitude())
	assert.Equal(t, assignedAt, order.AssignedAt())
}

func TestDeliveryOrder_Distance(t *testing.T) {
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	currentLoc := vo.MustNewLocation(52.5150, 13.4000)

	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())

	// Test distance calculations
	distToPickup := order.DistanceToPickup(currentLoc)
	distToDelivery := order.DistanceToDelivery(currentLoc)
	totalDist := order.TotalDistance()

	assert.Greater(t, distToPickup, 0.0)
	assert.Greater(t, distToDelivery, 0.0)
	assert.Greater(t, totalDist, 0.0)

	// Distance from pickup to delivery should match total distance
	assert.InDelta(t, pickup.DistanceTo(delivery), totalDist, 0.001)
}

func TestDeliverySimulator_StartDelivery(t *testing.T) {
	// Create a mock route generator
	routeGen, err := NewRouteGenerator(RouteGeneratorConfig{
		OSRMBaseURL: "http://localhost:5000",
		Timeout:     1 * time.Second,
	})
	require.NoError(t, err)

	defer routeGen.Close()

	locationPub := newMockLocationPublisher()
	statusPub := newMockStatusPublisher()

	config := DeliverySimulatorConfig{
		UpdateInterval:   100 * time.Millisecond,
		SpeedKmH:         100.0, // Fast speed for testing
		TimeMultiplier:   100.0, // Speed up time
		PickupWaitTime:   50 * time.Millisecond,
		DeliveryWaitTime: 50 * time.Millisecond,
		FailureRate:      0.0, // Always succeed
	}

	simulator := NewDeliverySimulator(config, routeGen, locationPub, statusPub)
	defer simulator.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a delivery order
	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5201, 13.4051) // Very close for fast test
	order := vo.NewDeliveryOrder("order-test-1", "pkg-test-1", pickup, delivery, time.Now())

	// Start the delivery
	err = simulator.StartDelivery(ctx, "courier-1", order)
	require.NoError(t, err)

	// Check initial state
	state, exists := simulator.GetDeliveryState("courier-1")
	require.True(t, exists)
	assert.Equal(t, "courier-1", state.CourierID)
	assert.Equal(t, vo.PhaseHeadingToPickup, state.Phase)
}

func TestDeliverySimulator_DoubleStartError(t *testing.T) {
	routeGen, err := NewRouteGenerator(RouteGeneratorConfig{
		OSRMBaseURL: "http://localhost:5000",
		Timeout:     1 * time.Second,
	})
	require.NoError(t, err)

	defer routeGen.Close()

	locationPub := newMockLocationPublisher()
	statusPub := newMockStatusPublisher()

	config := DefaultDeliverySimulatorConfig()

	simulator := NewDeliverySimulator(config, routeGen, locationPub, statusPub)
	defer simulator.Stop()

	ctx := context.Background()

	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())

	// First start should succeed
	err = simulator.StartDelivery(ctx, "courier-1", order)
	require.NoError(t, err)

	// Second start should fail (courier already active)
	order2 := vo.NewDeliveryOrder("order-2", "pkg-2", pickup, delivery, time.Now())
	err = simulator.StartDelivery(ctx, "courier-1", order2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already has an active delivery")
}

func TestDeliverySimulator_GetAllDeliveries(t *testing.T) {
	routeGen, err := NewRouteGenerator(RouteGeneratorConfig{
		OSRMBaseURL: "http://localhost:5000",
		Timeout:     1 * time.Second,
	})
	require.NoError(t, err)

	defer routeGen.Close()

	locationPub := newMockLocationPublisher()
	statusPub := newMockStatusPublisher()

	config := DefaultDeliverySimulatorConfig()

	simulator := NewDeliverySimulator(config, routeGen, locationPub, statusPub)
	defer simulator.Stop()

	ctx := context.Background()

	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)

	// Start multiple deliveries
	for i := 1; i <= 3; i++ {
		order := vo.NewDeliveryOrder("order-"+string(rune('0'+i)), "pkg-"+string(rune('0'+i)), pickup, delivery, time.Now())
		err = simulator.StartDelivery(ctx, "courier-"+string(rune('0'+i)), order)
		require.NoError(t, err)
	}

	// Check all deliveries
	ids := simulator.GetAllDeliveries()
	assert.Len(t, ids, 3)
}

func TestDeliverySimulator_StopDelivery(t *testing.T) {
	routeGen, err := NewRouteGenerator(RouteGeneratorConfig{
		OSRMBaseURL: "http://localhost:5000",
		Timeout:     1 * time.Second,
	})
	require.NoError(t, err)

	defer routeGen.Close()

	locationPub := newMockLocationPublisher()
	statusPub := newMockStatusPublisher()

	config := DefaultDeliverySimulatorConfig()

	simulator := NewDeliverySimulator(config, routeGen, locationPub, statusPub)
	defer simulator.Stop()

	ctx := context.Background()

	pickup := vo.MustNewLocation(52.5200, 13.4050)
	delivery := vo.MustNewLocation(52.5300, 13.4150)
	order := vo.NewDeliveryOrder("order-1", "pkg-1", pickup, delivery, time.Now())

	err = simulator.StartDelivery(ctx, "courier-1", order)
	require.NoError(t, err)

	// Verify it exists
	_, exists := simulator.GetDeliveryState("courier-1")
	assert.True(t, exists)

	// Stop the delivery
	simulator.StopDelivery("courier-1")

	// Verify it's gone
	_, exists = simulator.GetDeliveryState("courier-1")
	assert.False(t, exists)
}

func TestDefaultDeliverySimulatorConfig(t *testing.T) {
	config := DefaultDeliverySimulatorConfig()

	assert.Equal(t, 5*time.Second, config.UpdateInterval)
	assert.Equal(t, 30.0, config.SpeedKmH)
	assert.Equal(t, 1.0, config.TimeMultiplier)
	assert.Equal(t, 30*time.Second, config.PickupWaitTime)
	assert.Equal(t, 60*time.Second, config.DeliveryWaitTime)
	assert.Equal(t, 0.05, config.FailureRate)
}
