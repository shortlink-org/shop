package vo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoute_Valid(t *testing.T) {
	origin := MustNewLocation(52.5200, 13.4050)
	destination := MustNewLocation(52.5300, 13.4150)
	polyline := MustNewPolyline("_p~iF~ps|U_ulLnnqC")

	route, err := NewRoute("route-1", origin, destination, polyline, 1500.0, 3*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, "route-1", route.ID())
	assert.Equal(t, origin, route.Origin())
	assert.Equal(t, destination, route.Destination())
	assert.Equal(t, 1500.0, route.Distance())
	assert.Equal(t, 1.5, route.DistanceKm())
	assert.Equal(t, 3*time.Minute, route.Duration())
	assert.False(t, route.CreatedAt().IsZero())
}

func TestNewRoute_InvalidDistance(t *testing.T) {
	origin := MustNewLocation(52.5200, 13.4050)
	destination := MustNewLocation(52.5300, 13.4150)
	polyline := MustNewPolyline("_p~iF~ps|U_ulLnnqC")

	_, err := NewRoute("route-1", origin, destination, polyline, 0, 3*time.Minute)
	assert.ErrorIs(t, err, ErrInvalidDistance)

	_, err = NewRoute("route-1", origin, destination, polyline, -100, 3*time.Minute)
	assert.ErrorIs(t, err, ErrInvalidDistance)
}

func TestNewRoute_InvalidDuration(t *testing.T) {
	origin := MustNewLocation(52.5200, 13.4050)
	destination := MustNewLocation(52.5300, 13.4150)
	polyline := MustNewPolyline("_p~iF~ps|U_ulLnnqC")

	_, err := NewRoute("route-1", origin, destination, polyline, 1500, 0)
	assert.ErrorIs(t, err, ErrInvalidDuration)

	_, err = NewRoute("route-1", origin, destination, polyline, 1500, -time.Minute)
	assert.ErrorIs(t, err, ErrInvalidDuration)
}

func TestRoute_String(t *testing.T) {
	origin := MustNewLocation(52.5200, 13.4050)
	destination := MustNewLocation(52.5300, 13.4150)
	polyline := MustNewPolyline("_p~iF~ps|U_ulLnnqC")

	route, _ := NewRoute("route-1", origin, destination, polyline, 1500.0, 3*time.Minute)
	str := route.String()

	assert.Contains(t, str, "route-1")
	assert.Contains(t, str, "1.5 km")
	assert.Contains(t, str, "3m")
}
