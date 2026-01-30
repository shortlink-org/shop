package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolyline_Valid(t *testing.T) {
	// Example polyline from Berlin
	encoded := "_p~iF~ps|U_ulLnnqC_mqNvxq`@"
	p, err := NewPolyline(encoded)

	require.NoError(t, err)
	assert.Equal(t, encoded, p.Encoded())
}

func TestNewPolyline_Empty(t *testing.T) {
	_, err := NewPolyline("")
	assert.ErrorIs(t, err, ErrEmptyPolyline)
}

func TestPolyline_Decode(t *testing.T) {
	// This polyline represents a simple path: (38.5, -120.2), (40.7, -120.95), (43.252, -126.453)
	encoded := "_p~iF~ps|U_ulLnnqC_mqNvxq`@"

	p := MustNewPolyline(encoded)
	locations, err := p.Decode()

	require.NoError(t, err)
	require.Len(t, locations, 3)

	// Check first point (approximately)
	assert.InDelta(t, 38.5, locations[0].Latitude(), 0.01)
	assert.InDelta(t, -120.2, locations[0].Longitude(), 0.01)

	// Check second point (approximately)
	assert.InDelta(t, 40.7, locations[1].Latitude(), 0.01)
	assert.InDelta(t, -120.95, locations[1].Longitude(), 0.01)

	// Check third point (approximately)
	assert.InDelta(t, 43.252, locations[2].Latitude(), 0.01)
	assert.InDelta(t, -126.453, locations[2].Longitude(), 0.01)
}

func TestPolyline_String(t *testing.T) {
	encoded := "_p~iF~ps|U_ulLnnqC"
	p := MustNewPolyline(encoded)
	assert.Equal(t, encoded, p.String())
}

func TestPolyline_Len(t *testing.T) {
	encoded := "_p~iF~ps|U_ulLnnqC"
	p := MustNewPolyline(encoded)
	assert.Equal(t, len(encoded), p.Len())
}

func TestMustNewPolyline_Panics(t *testing.T) {
	assert.Panics(t, func() {
		MustNewPolyline("")
	})
}
