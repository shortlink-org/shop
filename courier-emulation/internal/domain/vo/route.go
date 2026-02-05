package vo

import (
	"errors"
	"fmt"
	"time"
)

// Route validation errors
var (
	ErrInvalidDistance = errors.New("distance must be positive")
	ErrInvalidDuration = errors.New("duration must be positive")
)

// Route represents a complete route as a value object.
// Contains origin, destination, polyline, distance and duration.
type Route struct {
	id          string
	origin      Location
	destination Location
	polyline    Polyline
	distance    float64       // in meters
	duration    time.Duration // estimated travel time
	createdAt   time.Time
}

// NewRoute creates a new Route value object with validation.
func NewRoute(
	id string,
	origin, destination Location,
	polyline Polyline,
	distanceMeters float64,
	duration time.Duration,
) (Route, error) {
	if distanceMeters <= 0 {
		return Route{}, fmt.Errorf("%w: %f", ErrInvalidDistance, distanceMeters)
	}

	if duration <= 0 {
		return Route{}, fmt.Errorf("%w: %v", ErrInvalidDuration, duration)
	}

	return Route{
		id:          id,
		origin:      origin,
		destination: destination,
		polyline:    polyline,
		distance:    distanceMeters,
		duration:    duration,
		createdAt:   time.Now(),
	}, nil
}

// ID returns the route identifier.
func (r Route) ID() string {
	return r.id
}

// Origin returns the starting location.
func (r Route) Origin() Location {
	return r.origin
}

// Destination returns the ending location.
func (r Route) Destination() Location {
	return r.destination
}

// Polyline returns the encoded polyline.
func (r Route) Polyline() Polyline {
	return r.polyline
}

// Distance returns the distance in meters.
func (r Route) Distance() float64 {
	return r.distance
}

// DistanceKm returns the distance in kilometers.
func (r Route) DistanceKm() float64 {
	return r.distance / 1000.0
}

// Duration returns the estimated travel time.
func (r Route) Duration() time.Duration {
	return r.duration
}

// CreatedAt returns when the route was created.
func (r Route) CreatedAt() time.Time {
	return r.createdAt
}

// Points decodes the polyline and returns all route points.
func (r Route) Points() ([]Location, error) {
	return r.polyline.Decode()
}

// String returns a string representation of the route.
func (r Route) String() string {
	return fmt.Sprintf("Route[%s]: %s -> %s (%.1f km, %v)",
		r.id, r.origin, r.destination, r.DistanceKm(), r.duration.Round(time.Second))
}
