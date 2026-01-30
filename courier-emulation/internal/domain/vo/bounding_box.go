package vo

import (
	"errors"
	"fmt"
	"math/rand"
)

// BoundingBox validation errors
var (
	ErrInvalidBoundingBox = errors.New("invalid bounding box: min must be less than max")
)

// BoundingBox represents a geographic bounding box as a value object.
// Used for generating random points within a region.
type BoundingBox struct {
	minLat float64
	maxLat float64
	minLon float64
	maxLon float64
}

// NewBoundingBox creates a new BoundingBox value object with validation.
func NewBoundingBox(minLat, maxLat, minLon, maxLon float64) (BoundingBox, error) {
	if minLat >= maxLat {
		return BoundingBox{}, fmt.Errorf("%w: minLat (%f) >= maxLat (%f)",
			ErrInvalidBoundingBox, minLat, maxLat)
	}
	if minLon >= maxLon {
		return BoundingBox{}, fmt.Errorf("%w: minLon (%f) >= maxLon (%f)",
			ErrInvalidBoundingBox, minLon, maxLon)
	}

	// Validate coordinates are within valid ranges
	if _, err := NewLocation(minLat, minLon); err != nil {
		return BoundingBox{}, err
	}
	if _, err := NewLocation(maxLat, maxLon); err != nil {
		return BoundingBox{}, err
	}

	return BoundingBox{
		minLat: minLat,
		maxLat: maxLat,
		minLon: minLon,
		maxLon: maxLon,
	}, nil
}

// MustNewBoundingBox creates a new BoundingBox or panics if invalid.
func MustNewBoundingBox(minLat, maxLat, minLon, maxLon float64) BoundingBox {
	bb, err := NewBoundingBox(minLat, maxLat, minLon, maxLon)
	if err != nil {
		panic(fmt.Sprintf("invalid bounding box: %v", err))
	}
	return bb
}

// BerlinBoundingBox returns the bounding box for Berlin, Germany.
func BerlinBoundingBox() BoundingBox {
	return MustNewBoundingBox(52.3383, 52.6755, 13.0884, 13.7610)
}

// MinLat returns the minimum latitude.
func (bb BoundingBox) MinLat() float64 {
	return bb.minLat
}

// MaxLat returns the maximum latitude.
func (bb BoundingBox) MaxLat() float64 {
	return bb.maxLat
}

// MinLon returns the minimum longitude.
func (bb BoundingBox) MinLon() float64 {
	return bb.minLon
}

// MaxLon returns the maximum longitude.
func (bb BoundingBox) MaxLon() float64 {
	return bb.maxLon
}

// RandomPoint generates a random location within the bounding box.
func (bb BoundingBox) RandomPoint() Location {
	lat := rand.Float64()*(bb.maxLat-bb.minLat) + bb.minLat
	lon := rand.Float64()*(bb.maxLon-bb.minLon) + bb.minLon
	return MustNewLocation(lat, lon)
}

// RandomPointPair generates two random locations within the bounding box.
// Useful for generating route start and end points.
func (bb BoundingBox) RandomPointPair() (Location, Location) {
	return bb.RandomPoint(), bb.RandomPoint()
}

// Contains checks if a location is within the bounding box.
func (bb BoundingBox) Contains(loc Location) bool {
	return loc.Latitude() >= bb.minLat &&
		loc.Latitude() <= bb.maxLat &&
		loc.Longitude() >= bb.minLon &&
		loc.Longitude() <= bb.maxLon
}

// String returns a string representation of the bounding box.
func (bb BoundingBox) String() string {
	return fmt.Sprintf("BoundingBox(lat: %.4f-%.4f, lon: %.4f-%.4f)",
		bb.minLat, bb.maxLat, bb.minLon, bb.maxLon)
}
