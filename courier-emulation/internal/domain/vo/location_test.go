package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocation_Valid(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{"Moscow", 55.7558, 37.6173},
		{"Berlin", 52.5200, 13.4050},
		{"Equator", 0.0, 0.0},
		{"North Pole", 90.0, 0.0},
		{"South Pole", -90.0, 0.0},
		{"Date Line East", 0.0, 180.0},
		{"Date Line West", 0.0, -180.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := NewLocation(tt.latitude, tt.longitude)
			require.NoError(t, err)
			assert.Equal(t, tt.latitude, loc.Latitude())
			assert.Equal(t, tt.longitude, loc.Longitude())
		})
	}
}

func TestNewLocation_InvalidLatitude(t *testing.T) {
	_, err := NewLocation(91.0, 0.0)
	assert.ErrorIs(t, err, ErrInvalidLatitude)

	_, err = NewLocation(-91.0, 0.0)
	assert.ErrorIs(t, err, ErrInvalidLatitude)
}

func TestNewLocation_InvalidLongitude(t *testing.T) {
	_, err := NewLocation(0.0, 181.0)
	assert.ErrorIs(t, err, ErrInvalidLongitude)

	_, err = NewLocation(0.0, -181.0)
	assert.ErrorIs(t, err, ErrInvalidLongitude)
}

func TestLocation_DistanceTo(t *testing.T) {
	moscow := MustNewLocation(55.7558, 37.6173)
	spb := MustNewLocation(59.9343, 30.3351)

	distance := moscow.DistanceTo(spb)

	// Distance between Moscow and St. Petersburg is approximately 635 km
	assert.InDelta(t, 635.0, distance, 50.0)
}

func TestLocation_DistanceTo_SamePoint(t *testing.T) {
	loc := MustNewLocation(52.5200, 13.4050)
	assert.Equal(t, 0.0, loc.DistanceTo(loc))
}

func TestLocation_String(t *testing.T) {
	loc := MustNewLocation(52.520008, 13.404954)
	assert.Equal(t, "(52.520008, 13.404954)", loc.String())
}

func TestLocation_ToOSRMFormat(t *testing.T) {
	loc := MustNewLocation(52.520008, 13.404954)
	// OSRM uses lon,lat format
	assert.Contains(t, loc.ToOSRMFormat(), "13.404954")
	assert.Contains(t, loc.ToOSRMFormat(), "52.520008")
}

func TestMustNewLocation_Panics(t *testing.T) {
	assert.Panics(t, func() {
		MustNewLocation(100.0, 0.0)
	})
}
