package services

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

// LocationPublisher interface for publishing location events.
type LocationPublisher interface {
	PublishLocation(ctx context.Context, event vo.CourierLocationEvent) error
	Close() error
}

// CourierState represents the current state of a simulated courier.
type CourierState struct {
	ID              string
	CurrentLocation vo.Location
	CurrentRoute    *vo.Route
	RoutePoints     []vo.Location
	CurrentPointIdx int
	Status          string
	Speed           float64 // km/h
	StartedAt       time.Time
	LastUpdateAt    time.Time

	// Delivery workflow fields
	CurrentOrder   *vo.DeliveryOrder // Current order being delivered
	Phase          vo.DeliveryPhase  // Current phase in delivery workflow
	PhaseStartedAt time.Time         // When current phase started
	PickupRoute    *vo.Route         // Route to pickup point
	DeliveryRoute  *vo.Route         // Route from pickup to customer
}

// CourierSimulatorConfig holds configuration for the courier simulator.
type CourierSimulatorConfig struct {
	UpdateInterval time.Duration // how often to publish location updates
	SpeedKmH       float64       // simulation speed in km/h
	TimeMultiplier float64       // time acceleration (1.0 = real-time, 2.0 = 2x speed)
}

// DefaultCourierSimulatorConfig returns default configuration.
func DefaultCourierSimulatorConfig() CourierSimulatorConfig {
	return CourierSimulatorConfig{
		UpdateInterval: 5 * time.Second,
		SpeedKmH:       30.0,
		TimeMultiplier: 1.0,
	}
}

// CourierSimulator simulates courier movement along routes.
type CourierSimulator struct {
	config         CourierSimulatorConfig
	routeGenerator *RouteGenerator
	publisher      LocationPublisher
	couriers       map[string]*CourierState
	mu             sync.RWMutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// NewCourierSimulator creates a new courier simulator.
func NewCourierSimulator(
	config CourierSimulatorConfig,
	routeGenerator *RouteGenerator,
	publisher LocationPublisher,
) *CourierSimulator {
	return &CourierSimulator{
		config:         config,
		routeGenerator: routeGenerator,
		publisher:      publisher,
		couriers:       make(map[string]*CourierState),
		stopCh:         make(chan struct{}),
	}
}

// StartCourier starts a new courier simulation with a random route.
func (cs *CourierSimulator) StartCourier(ctx context.Context, courierID string, bbox vo.BoundingBox) error {
	// Generate a random route
	route, err := cs.routeGenerator.GenerateRandomRoute(ctx, bbox)
	if err != nil {
		return fmt.Errorf("failed to generate route: %w", err)
	}

	return cs.StartCourierWithRoute(ctx, courierID, route)
}

// StartCourierWithRoute starts a courier simulation with a specific route.
func (cs *CourierSimulator) StartCourierWithRoute(ctx context.Context, courierID string, route vo.Route) error {
	// Decode polyline to get route points
	points, err := route.Points()
	if err != nil {
		return fmt.Errorf("failed to decode route points: %w", err)
	}

	if len(points) < 2 {
		return fmt.Errorf("route must have at least 2 points")
	}

	cs.mu.Lock()
	cs.couriers[courierID] = &CourierState{
		ID:              courierID,
		CurrentLocation: points[0],
		CurrentRoute:    &route,
		RoutePoints:     points,
		CurrentPointIdx: 0,
		Status:          vo.CourierStatusMoving,
		Speed:           cs.config.SpeedKmH,
		StartedAt:       time.Now(),
		LastUpdateAt:    time.Now(),
	}
	cs.mu.Unlock()

	// Start simulation goroutine
	cs.wg.Add(1)
	go cs.simulateCourier(ctx, courierID)

	return nil
}

// simulateCourier runs the simulation loop for a single courier.
func (cs *CourierSimulator) simulateCourier(ctx context.Context, courierID string) {
	defer cs.wg.Done()

	ticker := time.NewTicker(cs.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cs.stopCh:
			return
		case <-ticker.C:
			if err := cs.updateCourierPosition(ctx, courierID); err != nil {
				// Courier finished or error - stop simulation
				return
			}
		}
	}
}

// updateCourierPosition updates the courier's position and publishes the location.
func (cs *CourierSimulator) updateCourierPosition(ctx context.Context, courierID string) error {
	cs.mu.Lock()
	courier, exists := cs.couriers[courierID]
	if !exists {
		cs.mu.Unlock()
		return fmt.Errorf("courier %s not found", courierID)
	}

	// Calculate distance to travel based on time and speed
	elapsed := time.Since(courier.LastUpdateAt)
	distanceToTravel := (courier.Speed / 3600.0) * elapsed.Seconds() * cs.config.TimeMultiplier // km

	// Move along the route
	for distanceToTravel > 0 && courier.CurrentPointIdx < len(courier.RoutePoints)-1 {
		nextPoint := courier.RoutePoints[courier.CurrentPointIdx+1]
		distanceToNext := courier.CurrentLocation.DistanceTo(nextPoint)

		if distanceToTravel >= distanceToNext {
			// Move to next point
			courier.CurrentLocation = nextPoint
			courier.CurrentPointIdx++
			distanceToTravel -= distanceToNext
		} else {
			// Interpolate position
			ratio := distanceToTravel / distanceToNext
			courier.CurrentLocation = interpolateLocation(courier.CurrentLocation, nextPoint, ratio)
			distanceToTravel = 0
		}
	}

	// Check if route is completed
	if courier.CurrentPointIdx >= len(courier.RoutePoints)-1 {
		courier.Status = vo.CourierStatusIdle
		courier.CurrentLocation = courier.RoutePoints[len(courier.RoutePoints)-1]
	}

	courier.LastUpdateAt = time.Now()

	// Calculate heading
	heading := 0.0
	if courier.CurrentPointIdx < len(courier.RoutePoints)-1 {
		heading = calculateHeading(courier.CurrentLocation, courier.RoutePoints[courier.CurrentPointIdx+1])
	}

	// Create event
	event := vo.NewCourierLocationEvent(courierID, courier.CurrentLocation, courier.Status).
		WithSpeed(courier.Speed).
		WithHeading(heading).
		WithRouteID(courier.CurrentRoute.ID())

	isFinished := courier.Status == vo.CourierStatusIdle
	cs.mu.Unlock()

	// Publish location update
	if cs.publisher != nil {
		if err := cs.publisher.PublishLocation(ctx, event); err != nil {
			return fmt.Errorf("failed to publish location: %w", err)
		}
	}

	if isFinished {
		return fmt.Errorf("route completed")
	}

	return nil
}

// GetCourierState returns the current state of a courier.
func (cs *CourierSimulator) GetCourierState(courierID string) (*CourierState, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	state, exists := cs.couriers[courierID]
	if !exists {
		return nil, false
	}

	// Return a copy
	stateCopy := *state
	return &stateCopy, true
}

// GetAllCouriers returns all courier IDs.
func (cs *CourierSimulator) GetAllCouriers() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	ids := make([]string, 0, len(cs.couriers))
	for id := range cs.couriers {
		ids = append(ids, id)
	}
	return ids
}

// StopCourier stops a specific courier simulation.
func (cs *CourierSimulator) StopCourier(courierID string) {
	cs.mu.Lock()
	delete(cs.couriers, courierID)
	cs.mu.Unlock()
}

// Stop stops all courier simulations.
func (cs *CourierSimulator) Stop() {
	close(cs.stopCh)
	cs.wg.Wait()

	cs.mu.Lock()
	cs.couriers = make(map[string]*CourierState)
	cs.mu.Unlock()
}

// interpolateLocation calculates a point between two locations based on ratio (0-1).
func interpolateLocation(from, to vo.Location, ratio float64) vo.Location {
	lat := from.Latitude() + (to.Latitude()-from.Latitude())*ratio
	lon := from.Longitude() + (to.Longitude()-from.Longitude())*ratio
	return vo.MustNewLocation(lat, lon)
}

// calculateHeading calculates the heading (bearing) between two points in degrees.
func calculateHeading(from, to vo.Location) float64 {
	lat1 := from.Latitude() * math.Pi / 180
	lat2 := to.Latitude() * math.Pi / 180
	dLon := (to.Longitude() - from.Longitude()) * math.Pi / 180

	x := math.Sin(dLon) * math.Cos(lat2)
	y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)

	heading := math.Atan2(x, y) * 180 / math.Pi
	if heading < 0 {
		heading += 360
	}
	return heading
}
