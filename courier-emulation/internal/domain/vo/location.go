package vo

import (
	"errors"
	"fmt"
	"math"
)

// Location validation errors
var (
	ErrInvalidLatitude  = errors.New("latitude is out of valid range")
	ErrInvalidLongitude = errors.New("longitude is out of valid range")
)

// Constants for location validation bounds
const (
	MinLatitude  float64 = -90.0
	MaxLatitude  float64 = 90.0
	MinLongitude float64 = -180.0
	MaxLongitude float64 = 180.0

	// EarthRadiusKm is the Earth's radius in kilometers
	EarthRadiusKm float64 = 6371.0
)

// Location represents a GPS location as a value object.
// A value object is immutable and defined by its attributes.
type Location struct {
	latitude  float64
	longitude float64
}

// NewLocation creates a new Location value object with validation.
func NewLocation(latitude, longitude float64) (Location, error) {
	if latitude < MinLatitude || latitude > MaxLatitude {
		return Location{}, fmt.Errorf("%w: %f (must be between %f and %f)",
			ErrInvalidLatitude, latitude, MinLatitude, MaxLatitude)
	}

	if longitude < MinLongitude || longitude > MaxLongitude {
		return Location{}, fmt.Errorf("%w: %f (must be between %f and %f)",
			ErrInvalidLongitude, longitude, MinLongitude, MaxLongitude)
	}

	return Location{
		latitude:  latitude,
		longitude: longitude,
	}, nil
}

// MustNewLocation creates a new Location or panics if invalid.
func MustNewLocation(latitude, longitude float64) Location {
	loc, err := NewLocation(latitude, longitude)
	if err != nil {
		panic(fmt.Sprintf("invalid location: %v", err))
	}

	return loc
}

// Latitude returns the latitude in degrees.
func (l Location) Latitude() float64 {
	return l.latitude
}

// Longitude returns the longitude in degrees.
func (l Location) Longitude() float64 {
	return l.longitude
}

// DistanceTo calculates the distance to another location using Haversine formula.
// Returns distance in kilometers.
func (l Location) DistanceTo(other Location) float64 {
	lat1Rad := l.latitude * math.Pi / 180
	lat2Rad := other.latitude * math.Pi / 180
	deltaLat := (other.latitude - l.latitude) * math.Pi / 180
	deltaLon := (other.longitude - l.longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

// String returns a string representation of the location.
func (l Location) String() string {
	return fmt.Sprintf("(%.6f, %.6f)", l.latitude, l.longitude)
}

// ToOSRMFormat returns the location in OSRM API format (lon,lat).
func (l Location) ToOSRMFormat() string {
	return fmt.Sprintf("%f,%f", l.longitude, l.latitude)
}
