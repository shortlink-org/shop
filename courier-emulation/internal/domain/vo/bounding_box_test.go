package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBoundingBox_Valid(t *testing.T) {
	bb, err := NewBoundingBox(52.3383, 52.6755, 13.0884, 13.7610)

	require.NoError(t, err)
	assert.Equal(t, 52.3383, bb.MinLat())
	assert.Equal(t, 52.6755, bb.MaxLat())
	assert.Equal(t, 13.0884, bb.MinLon())
	assert.Equal(t, 13.7610, bb.MaxLon())
}

func TestNewBoundingBox_InvalidLatRange(t *testing.T) {
	_, err := NewBoundingBox(52.6755, 52.3383, 13.0884, 13.7610)
	assert.ErrorIs(t, err, ErrInvalidBoundingBox)
}

func TestNewBoundingBox_InvalidLonRange(t *testing.T) {
	_, err := NewBoundingBox(52.3383, 52.6755, 13.7610, 13.0884)
	assert.ErrorIs(t, err, ErrInvalidBoundingBox)
}

func TestBerlinBoundingBox(t *testing.T) {
	bb := BerlinBoundingBox()

	assert.InDelta(t, 52.3383, bb.MinLat(), 0.0001)
	assert.InDelta(t, 52.6755, bb.MaxLat(), 0.0001)
	assert.InDelta(t, 13.0884, bb.MinLon(), 0.0001)
	assert.InDelta(t, 13.7610, bb.MaxLon(), 0.0001)
}

func TestBoundingBox_RandomPoint(t *testing.T) {
	bb := BerlinBoundingBox()

	for range 100 {
		point := bb.RandomPoint()
		assert.True(t, bb.Contains(point), "point %v should be within bounding box", point)
	}
}

func TestBoundingBox_RandomPointPair(t *testing.T) {
	bb := BerlinBoundingBox()

	for range 50 {
		p1, p2 := bb.RandomPointPair()
		assert.True(t, bb.Contains(p1))
		assert.True(t, bb.Contains(p2))
	}
}

func TestBoundingBox_Contains(t *testing.T) {
	bb := BerlinBoundingBox()

	// Berlin center should be contained
	berlin := MustNewLocation(52.5200, 13.4050)
	assert.True(t, bb.Contains(berlin))

	// Moscow should not be contained
	moscow := MustNewLocation(55.7558, 37.6173)
	assert.False(t, bb.Contains(moscow))
}

func TestBoundingBox_String(t *testing.T) {
	bb := BerlinBoundingBox()
	str := bb.String()

	assert.Contains(t, str, "BoundingBox")
	assert.Contains(t, str, "52.3383")
	assert.Contains(t, str, "13.7610")
}
