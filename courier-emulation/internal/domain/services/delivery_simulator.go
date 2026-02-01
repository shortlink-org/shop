package services

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/infrastructure/kafka"
)

// DeliverySimulatorConfig holds configuration for the delivery simulator.
type DeliverySimulatorConfig struct {
	UpdateInterval   time.Duration // How often to update courier position
	SpeedKmH         float64       // Courier speed in km/h
	TimeMultiplier   float64       // Time acceleration (1.0 = real-time)
	PickupWaitTime   time.Duration // Time to wait at pickup location
	DeliveryWaitTime time.Duration // Time to wait at delivery location
	FailureRate      float64       // Probability of NOT_DELIVERED (0.0 - 1.0)
}

// DefaultDeliverySimulatorConfig returns default configuration.
func DefaultDeliverySimulatorConfig() DeliverySimulatorConfig {
	return DeliverySimulatorConfig{
		UpdateInterval:   5 * time.Second,
		SpeedKmH:         30.0,
		TimeMultiplier:   1.0,
		PickupWaitTime:   30 * time.Second,
		DeliveryWaitTime: 60 * time.Second,
		FailureRate:      0.05,
	}
}

// DeliveryState represents the current state of a delivery simulation.
type DeliveryState struct {
	CourierID       string
	CurrentLocation vo.Location
	CurrentOrder    *vo.DeliveryOrder
	Phase           vo.DeliveryPhase
	PhaseStartedAt  time.Time
	CurrentRoute    *vo.Route
	RoutePoints     []vo.Location
	CurrentPointIdx int
	Speed           float64
	LastUpdateAt    time.Time
}

// DeliverySimulator orchestrates the full delivery workflow simulation.
type DeliverySimulator struct {
	config          DeliverySimulatorConfig
	routeGenerator  *RouteGenerator
	locationPub     LocationPublisher
	statusPub       kafka.StatusPublisher
	deliveries      map[string]*DeliveryState
	mu              sync.RWMutex
	stopCh          chan struct{}
	wg              sync.WaitGroup
	rng             *rand.Rand
}

// NewDeliverySimulator creates a new delivery simulator.
func NewDeliverySimulator(
	config DeliverySimulatorConfig,
	routeGenerator *RouteGenerator,
	locationPub LocationPublisher,
	statusPub kafka.StatusPublisher,
) *DeliverySimulator {
	return &DeliverySimulator{
		config:         config,
		routeGenerator: routeGenerator,
		locationPub:    locationPub,
		statusPub:      statusPub,
		deliveries:     make(map[string]*DeliveryState),
		stopCh:         make(chan struct{}),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// StartDelivery starts a delivery simulation for a courier with an assigned order.
func (ds *DeliverySimulator) StartDelivery(ctx context.Context, courierID string, order vo.DeliveryOrder) error {
	ds.mu.Lock()

	// Check if courier already has an active delivery
	if existing, exists := ds.deliveries[courierID]; exists && existing.Phase != vo.PhaseIdle {
		ds.mu.Unlock()
		return fmt.Errorf("courier %s already has an active delivery", courierID)
	}

	ds.mu.Unlock()

	// Generate route to pickup location
	// For simplicity, we'll assume courier starts at pickup location or use a default starting point
	// In a real scenario, we'd get the courier's current location
	startLocation := order.PickupLocation()

	route, err := ds.routeGenerator.GenerateRoute(ctx, startLocation, order.PickupLocation())
	if err != nil {
		// If same location, create a minimal route
		route, _ = ds.createMinimalRoute(startLocation, order.PickupLocation())
	}

	points, err := route.Points()
	if err != nil || len(points) < 2 {
		// Create direct route
		points = []vo.Location{startLocation, order.PickupLocation()}
	}

	orderCopy := order
	state := &DeliveryState{
		CourierID:       courierID,
		CurrentLocation: points[0],
		CurrentOrder:    &orderCopy,
		Phase:           vo.PhaseHeadingToPickup,
		PhaseStartedAt:  time.Now(),
		CurrentRoute:    &route,
		RoutePoints:     points,
		CurrentPointIdx: 0,
		Speed:           ds.config.SpeedKmH,
		LastUpdateAt:    time.Now(),
	}

	ds.mu.Lock()
	ds.deliveries[courierID] = state
	ds.mu.Unlock()

	// Start simulation goroutine
	ds.wg.Add(1)
	go ds.simulateDelivery(ctx, courierID)

	return nil
}

// createMinimalRoute creates a minimal route between two points.
func (ds *DeliverySimulator) createMinimalRoute(from, to vo.Location) (vo.Route, error) {
	// Create a simple polyline with just two points
	polyline := vo.MustNewPolyline(encodePolyline([]vo.Location{from, to}))
	distance := from.DistanceTo(to)
	duration := time.Duration(distance/ds.config.SpeedKmH*3600) * time.Second

	return vo.NewRoute(
		fmt.Sprintf("minimal_%d", time.Now().UnixNano()),
		from,
		to,
		polyline,
		distance*1000, // Convert to meters
		duration,
	)
}

// encodePolyline is a simple polyline encoder for two points.
func encodePolyline(points []vo.Location) string {
	if len(points) == 0 {
		return ""
	}

	// Simple polyline encoding
	result := ""
	prevLat, prevLon := int64(0), int64(0)

	for _, p := range points {
		lat := int64(p.Latitude() * 1e5)
		lon := int64(p.Longitude() * 1e5)

		result += encodeNumber(lat - prevLat)
		result += encodeNumber(lon - prevLon)

		prevLat, prevLon = lat, lon
	}

	return result
}

func encodeNumber(num int64) string {
	if num < 0 {
		num = ^(num << 1)
	} else {
		num = num << 1
	}

	result := ""
	for num >= 0x20 {
		result += string(rune((0x20 | (num & 0x1f)) + 63))
		num >>= 5
	}
	result += string(rune(num + 63))

	return result
}

// simulateDelivery runs the simulation loop for a delivery.
func (ds *DeliverySimulator) simulateDelivery(ctx context.Context, courierID string) {
	defer ds.wg.Done()

	ticker := time.NewTicker(ds.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ds.stopCh:
			return
		case <-ticker.C:
			finished, err := ds.updateDelivery(ctx, courierID)
			if err != nil || finished {
				return
			}
		}
	}
}

// updateDelivery updates the delivery state and handles phase transitions.
func (ds *DeliverySimulator) updateDelivery(ctx context.Context, courierID string) (bool, error) {
	ds.mu.Lock()
	state, exists := ds.deliveries[courierID]
	if !exists {
		ds.mu.Unlock()
		return true, fmt.Errorf("delivery for courier %s not found", courierID)
	}

	// Handle based on current phase
	switch state.Phase {
	case vo.PhaseHeadingToPickup, vo.PhaseHeadingToCustomer:
		return ds.handleMovingPhase(ctx, state)

	case vo.PhasePickingUp:
		return ds.handlePickingUpPhase(ctx, state)

	case vo.PhaseDelivering:
		return ds.handleDeliveringPhase(ctx, state)

	case vo.PhaseIdle:
		ds.mu.Unlock()
		return true, nil

	default:
		ds.mu.Unlock()
		return true, fmt.Errorf("unknown phase: %s", state.Phase)
	}
}

// handleMovingPhase handles courier movement along a route.
func (ds *DeliverySimulator) handleMovingPhase(ctx context.Context, state *DeliveryState) (bool, error) {
	// Calculate distance to travel
	elapsed := time.Since(state.LastUpdateAt)
	distanceToTravel := (state.Speed / 3600.0) * elapsed.Seconds() * ds.config.TimeMultiplier // km

	// Move along the route
	for distanceToTravel > 0 && state.CurrentPointIdx < len(state.RoutePoints)-1 {
		nextPoint := state.RoutePoints[state.CurrentPointIdx+1]
		distanceToNext := state.CurrentLocation.DistanceTo(nextPoint)

		if distanceToTravel >= distanceToNext {
			state.CurrentLocation = nextPoint
			state.CurrentPointIdx++
			distanceToTravel -= distanceToNext
		} else {
			ratio := distanceToTravel / distanceToNext
			state.CurrentLocation = interpolateLocation(state.CurrentLocation, nextPoint, ratio)
			distanceToTravel = 0
		}
	}

	state.LastUpdateAt = time.Now()

	// Check if route is completed
	routeCompleted := state.CurrentPointIdx >= len(state.RoutePoints)-1
	if routeCompleted {
		state.CurrentLocation = state.RoutePoints[len(state.RoutePoints)-1]
	}

	// Calculate heading
	heading := 0.0
	if state.CurrentPointIdx < len(state.RoutePoints)-1 {
		heading = calculateHeading(state.CurrentLocation, state.RoutePoints[state.CurrentPointIdx+1])
	}

	// Create and publish location event
	event := vo.NewCourierLocationEvent(state.CourierID, state.CurrentLocation, state.Phase.ToCourierStatus()).
		WithSpeed(state.Speed).
		WithHeading(heading)

	if state.CurrentRoute != nil {
		event = event.WithRouteID(state.CurrentRoute.ID())
	}

	ds.mu.Unlock()

	// Publish location update
	if ds.locationPub != nil {
		if err := ds.locationPub.PublishLocation(ctx, event); err != nil {
			return false, fmt.Errorf("failed to publish location: %w", err)
		}
	}

	// Handle phase transition if route completed
	if routeCompleted {
		return ds.transitionPhase(ctx, state.CourierID)
	}

	return false, nil
}

// handlePickingUpPhase handles the pickup waiting phase.
func (ds *DeliverySimulator) handlePickingUpPhase(ctx context.Context, state *DeliveryState) (bool, error) {
	waitTime := time.Since(state.PhaseStartedAt) * time.Duration(ds.config.TimeMultiplier)

	// Publish stationary location update
	event := vo.NewCourierLocationEvent(state.CourierID, state.CurrentLocation, vo.CourierStatusPickingUp).
		WithSpeed(0)

	ds.mu.Unlock()

	if ds.locationPub != nil {
		if err := ds.locationPub.PublishLocation(ctx, event); err != nil {
			return false, fmt.Errorf("failed to publish location: %w", err)
		}
	}

	// Check if wait time is complete
	if waitTime >= ds.config.PickupWaitTime {
		return ds.transitionPhase(ctx, state.CourierID)
	}

	return false, nil
}

// handleDeliveringPhase handles the delivery waiting phase.
func (ds *DeliverySimulator) handleDeliveringPhase(ctx context.Context, state *DeliveryState) (bool, error) {
	waitTime := time.Since(state.PhaseStartedAt) * time.Duration(ds.config.TimeMultiplier)

	// Publish stationary location update
	event := vo.NewCourierLocationEvent(state.CourierID, state.CurrentLocation, vo.CourierStatusDelivering).
		WithSpeed(0)

	ds.mu.Unlock()

	if ds.locationPub != nil {
		if err := ds.locationPub.PublishLocation(ctx, event); err != nil {
			return false, fmt.Errorf("failed to publish location: %w", err)
		}
	}

	// Check if wait time is complete
	if waitTime >= ds.config.DeliveryWaitTime {
		return ds.transitionPhase(ctx, state.CourierID)
	}

	return false, nil
}

// transitionPhase handles phase transitions.
func (ds *DeliverySimulator) transitionPhase(ctx context.Context, courierID string) (bool, error) {
	ds.mu.Lock()
	state, exists := ds.deliveries[courierID]
	if !exists {
		ds.mu.Unlock()
		return true, fmt.Errorf("delivery not found")
	}

	currentPhase := state.Phase
	order := state.CurrentOrder

	switch currentPhase {
	case vo.PhaseHeadingToPickup:
		// Arrived at pickup -> start picking up
		state.Phase = vo.PhasePickingUp
		state.PhaseStartedAt = time.Now()
		ds.mu.Unlock()
		return false, nil

	case vo.PhasePickingUp:
		// Pickup complete -> publish event and generate route to customer
		ds.mu.Unlock()

		// Publish pickup event
		if ds.statusPub != nil && order != nil {
			pickupEvent := kafka.NewPickUpOrderEvent(courierID, *order, state.CurrentLocation)
			if err := ds.statusPub.PublishPickUp(ctx, pickupEvent); err != nil {
				return false, fmt.Errorf("failed to publish pickup event: %w", err)
			}
		}

		// Generate route to customer
		if order != nil {
			route, err := ds.routeGenerator.GenerateRoute(ctx, state.CurrentLocation, order.DeliveryLocation())
			if err != nil {
				route, _ = ds.createMinimalRoute(state.CurrentLocation, order.DeliveryLocation())
			}

			points, _ := route.Points()
			if len(points) < 2 {
				points = []vo.Location{state.CurrentLocation, order.DeliveryLocation()}
			}

			ds.mu.Lock()
			state.CurrentRoute = &route
			state.RoutePoints = points
			state.CurrentPointIdx = 0
			state.Phase = vo.PhaseHeadingToCustomer
			state.PhaseStartedAt = time.Now()
			state.LastUpdateAt = time.Now()
			ds.mu.Unlock()
		}

		return false, nil

	case vo.PhaseHeadingToCustomer:
		// Arrived at customer -> start delivering
		state.Phase = vo.PhaseDelivering
		state.PhaseStartedAt = time.Now()
		ds.mu.Unlock()
		return false, nil

	case vo.PhaseDelivering:
		// Delivery complete -> publish event and return to idle
		ds.mu.Unlock()

		// Determine if delivery was successful (based on failure rate)
		delivered := ds.rng.Float64() >= ds.config.FailureRate
		reason := ""
		if !delivered {
			reasons := []string{
				kafka.ReasonCustomerNotAvailable,
				kafka.ReasonWrongAddress,
				kafka.ReasonCustomerRefused,
				kafka.ReasonAccessDenied,
			}
			reason = reasons[ds.rng.Intn(len(reasons))]
		}

		// Publish delivery event
		if ds.statusPub != nil && order != nil {
			deliverEvent := kafka.NewDeliverOrderEvent(courierID, *order, state.CurrentLocation, delivered, reason)
			if err := ds.statusPub.PublishDelivery(ctx, deliverEvent); err != nil {
				return false, fmt.Errorf("failed to publish delivery event: %w", err)
			}
		}

		// Reset state to idle
		ds.mu.Lock()
		state.Phase = vo.PhaseIdle
		state.CurrentOrder = nil
		state.CurrentRoute = nil
		state.RoutePoints = nil
		state.CurrentPointIdx = 0
		ds.mu.Unlock()

		return true, nil

	default:
		ds.mu.Unlock()
		return true, nil
	}
}

// GetDeliveryState returns the current state of a delivery.
func (ds *DeliverySimulator) GetDeliveryState(courierID string) (*DeliveryState, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	state, exists := ds.deliveries[courierID]
	if !exists {
		return nil, false
	}

	// Return a copy
	stateCopy := *state
	return &stateCopy, true
}

// GetAllDeliveries returns all active delivery courier IDs.
func (ds *DeliverySimulator) GetAllDeliveries() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	ids := make([]string, 0, len(ds.deliveries))
	for id, state := range ds.deliveries {
		if state.Phase != vo.PhaseIdle {
			ids = append(ids, id)
		}
	}
	return ids
}

// StopDelivery stops a specific delivery simulation.
func (ds *DeliverySimulator) StopDelivery(courierID string) {
	ds.mu.Lock()
	delete(ds.deliveries, courierID)
	ds.mu.Unlock()
}

// Stop stops all delivery simulations.
func (ds *DeliverySimulator) Stop() {
	close(ds.stopCh)
	ds.wg.Wait()

	ds.mu.Lock()
	ds.deliveries = make(map[string]*DeliveryState)
	ds.mu.Unlock()
}
