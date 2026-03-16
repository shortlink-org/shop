package vo

import (
	"errors"
	"fmt"
)

const (
	// polylineASCIIShift converts encoded bytes to/from printable ASCII.
	polylineASCIIShift = 63
	// polylineChunkMask extracts the 5-bit payload from each encoded chunk.
	polylineChunkMask = 0x1f
	// polylineContinuationMask marks chunks that have more bytes following.
	polylineContinuationMask = 0x20
	// polylinePrecision is the fixed scaling factor used by Google polyline encoding.
	polylinePrecision = 1e5
)

// Polyline validation errors
var (
	ErrEmptyPolyline = errors.New("polyline cannot be empty")
)

// Polyline represents an encoded polyline string as a value object.
// Uses Google's Polyline Algorithm for encoding coordinates.
// https://developers.google.com/maps/documentation/utilities/polylinealgorithm
type Polyline struct {
	encoded string
}

// NewPolyline creates a new Polyline value object with validation.
func NewPolyline(encoded string) (Polyline, error) {
	if encoded == "" {
		return Polyline{}, ErrEmptyPolyline
	}

	return Polyline{
		encoded: encoded,
	}, nil
}

// MustNewPolyline creates a new Polyline or panics if invalid.
func MustNewPolyline(encoded string) Polyline {
	p, err := NewPolyline(encoded)
	if err != nil {
		panic(fmt.Sprintf("invalid polyline: %v", err))
	}

	return p
}

// Encoded returns the encoded polyline string.
func (p Polyline) Encoded() string {
	return p.encoded
}

// Decode decodes the polyline into a slice of Location points.
// Uses the Google Polyline Algorithm.
func (p Polyline) Decode() ([]Location, error) {
	if p.encoded == "" {
		return nil, ErrEmptyPolyline
	}

	var locations []Location

	index := 0
	lat := 0
	lng := 0

	for index < len(p.encoded) {
		// Decode latitude
		shift := 0
		result := 0

		for index < len(p.encoded) {
			b := int(p.encoded[index]) - polylineASCIIShift
			index++
			result |= (b & polylineChunkMask) << shift
			shift += 5

			if b < polylineContinuationMask {
				break
			}
		}

		if result&1 != 0 {
			lat += ^(result >> 1)
		} else {
			lat += result >> 1
		}

		// Decode longitude
		shift = 0
		result = 0

		for index < len(p.encoded) {
			b := int(p.encoded[index]) - polylineASCIIShift
			index++
			result |= (b & polylineChunkMask) << shift
			shift += 5

			if b < polylineContinuationMask {
				break
			}
		}

		if result&1 != 0 {
			lng += ^(result >> 1)
		} else {
			lng += result >> 1
		}

		// Convert to degrees (polyline uses 1e5 precision)
		latitude := float64(lat) / polylinePrecision
		longitude := float64(lng) / polylinePrecision

		loc, err := NewLocation(latitude, longitude)
		if err != nil {
			return nil, fmt.Errorf("invalid coordinate in polyline: %w", err)
		}

		locations = append(locations, loc)
	}

	return locations, nil
}

// String returns the encoded polyline string.
func (p Polyline) String() string {
	return p.encoded
}

// Len returns the length of the encoded string.
func (p Polyline) Len() int {
	return len(p.encoded)
}
