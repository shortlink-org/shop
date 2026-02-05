package location

import (
	"errors"
	"fmt"
)

// Location validation errors
var (
	ErrLocationInvalidLatitude  = errors.New("location latitude is out of valid range")
	ErrLocationInvalidLongitude = errors.New("location longitude is out of valid range")
)

// Constants for location validation bounds
const (
	// MinLatitude is the minimum valid latitude (-90.0)
	MinLatitude float64 = -90.0
	// MaxLatitude is the maximum valid latitude (90.0)
	MaxLatitude float64 = 90.0
	// MinLongitude is the minimum valid longitude (-180.0)
	MinLongitude float64 = -180.0
	// MaxLongitude is the maximum valid longitude (180.0)
	MaxLongitude float64 = 180.0
)

// Location represents a GPS location as a value object.
// A value object is immutable and defined by its attributes.
// Two locations are considered equal if they have the same latitude and longitude.
type Location struct {
	// latitude is the latitude coordinate in degrees
	latitude float64
	// longitude is the longitude coordinate in degrees
	longitude float64
}

// NewLocation creates a new Location value object with validation.
//
// Args:
//   - latitude: Latitude in degrees (MinLatitude to MaxLatitude)
//   - longitude: Longitude in degrees (MinLongitude to MaxLongitude)
//
// Returns:
//   - Location: The validated location value object
//   - error: Error if location is invalid
//
// Example:
//
//	loc, err := vo.NewLocation(55.7558, 37.6173) // Moscow coordinates
//	if err != nil {
//	    return err
//	}
func NewLocation(latitude, longitude float64) (Location, error) {
	if latitude < MinLatitude || latitude > MaxLatitude {
		return Location{}, fmt.Errorf("%w: %f (must be between %f and %f)", ErrLocationInvalidLatitude, latitude, MinLatitude, MaxLatitude)
	}

	if longitude < MinLongitude || longitude > MaxLongitude {
		return Location{}, fmt.Errorf("%w: %f (must be between %f and %f)", ErrLocationInvalidLongitude, longitude, MinLongitude, MaxLongitude)
	}

	return Location{
		latitude:  latitude,
		longitude: longitude,
	}, nil
}

// MustNewLocation creates a new Location value object or panics if invalid.
// Use this only when you are certain the location is valid.
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

// IsZero checks if the location is at zero coordinates (0, 0).
func (l Location) IsZero() bool {
	return l.latitude == 0 && l.longitude == 0
}

// String returns a string representation of the location.
func (l Location) String() string {
	return fmt.Sprintf("(%.6f, %.6f)", l.latitude, l.longitude)
}
