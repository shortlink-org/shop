package vo

import (
	"encoding/json"
	"time"
)

// CourierLocationEvent represents a courier's location update event.
// This is published to Kafka for real-time tracking.
type CourierLocationEvent struct {
	CourierID string    `json:"courier_id"`
	Location  Location  `json:"location"`
	Timestamp time.Time `json:"timestamp"`
	Speed     float64   `json:"speed_kmh,omitempty"`  // current speed in km/h
	Heading   float64   `json:"heading,omitempty"`    // heading in degrees (0-360)
	RouteID   string    `json:"route_id,omitempty"`   // current route being followed
	Status    string    `json:"status"`               // moving, idle, delivering
}

// NewCourierLocationEvent creates a new courier location event.
func NewCourierLocationEvent(courierID string, location Location, status string) CourierLocationEvent {
	return CourierLocationEvent{
		CourierID: courierID,
		Location:  location,
		Timestamp: time.Now(),
		Status:    status,
	}
}

// WithSpeed sets the speed for the event.
func (e CourierLocationEvent) WithSpeed(speedKmh float64) CourierLocationEvent {
	e.Speed = speedKmh
	return e
}

// WithHeading sets the heading for the event.
func (e CourierLocationEvent) WithHeading(heading float64) CourierLocationEvent {
	e.Heading = heading
	return e
}

// WithRouteID sets the route ID for the event.
func (e CourierLocationEvent) WithRouteID(routeID string) CourierLocationEvent {
	e.RouteID = routeID
	return e
}

// MarshalJSON implements custom JSON marshaling for Location.
func (e CourierLocationEvent) MarshalJSON() ([]byte, error) {
	type Alias CourierLocationEvent
	return json.Marshal(&struct {
		Alias
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}{
		Alias:     Alias(e),
		Latitude:  e.Location.Latitude(),
		Longitude: e.Location.Longitude(),
	})
}

// ToJSON serializes the event to JSON bytes.
func (e CourierLocationEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// CourierStatus constants
const (
	CourierStatusMoving     = "moving"
	CourierStatusIdle       = "idle"
	CourierStatusDelivering = "delivering"
	CourierStatusPickingUp  = "picking_up"
)
