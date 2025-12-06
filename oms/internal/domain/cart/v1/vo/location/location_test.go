package location

import (
	"errors"
	"strings"
	"testing"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		wantErr   bool
		errType   error
	}{
		{
			name:      "valid location - Moscow",
			latitude:  55.7558,
			longitude: 37.6173,
			wantErr:   false,
		},
		{
			name:      "valid location - zero coordinates",
			latitude:  0.0,
			longitude: 0.0,
			wantErr:   false,
		},
		{
			name:      "valid location - boundary values",
			latitude:  90.0,
			longitude: 180.0,
			wantErr:   false,
		},
		{
			name:      "invalid latitude too high",
			latitude:  91.0,
			longitude: 37.6173,
			wantErr:   true,
			errType:   ErrLocationInvalidLatitude,
		},
		{
			name:      "invalid latitude too low",
			latitude:  -91.0,
			longitude: 37.6173,
			wantErr:   true,
			errType:   ErrLocationInvalidLatitude,
		},
		{
			name:      "invalid longitude too high",
			latitude:  55.7558,
			longitude: 181.0,
			wantErr:   true,
			errType:   ErrLocationInvalidLongitude,
		},
		{
			name:      "invalid longitude too low",
			latitude:  55.7558,
			longitude: -181.0,
			wantErr:   true,
			errType:   ErrLocationInvalidLongitude,
		},
		{
			name:      "boundary values - minimum",
			latitude:  -90.0,
			longitude: -180.0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := NewLocation(tt.latitude, tt.longitude)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLocation() expected error but got none")
					return
				}
				if tt.errType != nil {
					if err != tt.errType && !errors.Is(err, tt.errType) {
						t.Errorf("NewLocation() error = %v, want %v", err, tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("NewLocation() unexpected error: %v", err)
					return
				}
				if loc.Latitude() != tt.latitude {
					t.Errorf("NewLocation() latitude = %v, want %v", loc.Latitude(), tt.latitude)
				}
				if loc.Longitude() != tt.longitude {
					t.Errorf("NewLocation() longitude = %v, want %v", loc.Longitude(), tt.longitude)
				}
			}
		})
	}
}

func TestLocation_Getters(t *testing.T) {
	latitude := 55.7558
	longitude := 37.6173

	loc, err := NewLocation(latitude, longitude)
	if err != nil {
		t.Fatalf("NewLocation() error = %v", err)
	}

	if loc.Latitude() != latitude {
		t.Errorf("Location.Latitude() = %v, want %v", loc.Latitude(), latitude)
	}
	if loc.Longitude() != longitude {
		t.Errorf("Location.Longitude() = %v, want %v", loc.Longitude(), longitude)
	}
}

func TestLocation_IsZero(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		want      bool
	}{
		{
			name:      "zero coordinates",
			latitude:  0.0,
			longitude: 0.0,
			want:      true,
		},
		{
			name:      "non-zero coordinates",
			latitude:  55.7558,
			longitude: 37.6173,
			want:      false,
		},
		{
			name:      "zero latitude only",
			latitude:  0.0,
			longitude: 37.6173,
			want:      false,
		},
		{
			name:      "zero longitude only",
			latitude:  55.7558,
			longitude: 0.0,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := NewLocation(tt.latitude, tt.longitude)
			if err != nil {
				t.Fatalf("NewLocation() error = %v", err)
			}
			if loc.IsZero() != tt.want {
				t.Errorf("Location.IsZero() = %v, want %v", loc.IsZero(), tt.want)
			}
		})
	}
}

func TestLocation_String(t *testing.T) {
	loc, err := NewLocation(55.7558, 37.6173)
	if err != nil {
		t.Fatalf("NewLocation() error = %v", err)
	}

	str := loc.String()
	if str == "" {
		t.Errorf("Location.String() should not be empty")
	}
	if !strings.Contains(str, "55.7558") || !strings.Contains(str, "37.6173") {
		t.Errorf("Location.String() should contain coordinates")
	}
}

func TestLocation_Equality(t *testing.T) {
	loc1, _ := NewLocation(55.7558, 37.6173)
	loc2, _ := NewLocation(55.7558, 37.6173)
	loc3, _ := NewLocation(59.9343, 30.3351)

	if loc1 != loc2 {
		t.Errorf("Locations with same coordinates should be equal")
	}
	if loc1 == loc3 {
		t.Errorf("Locations with different coordinates should not be equal")
	}
}

func TestMustNewLocation(t *testing.T) {
	t.Run("valid location", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewLocation() panicked on valid location: %v", r)
			}
		}()
		loc := MustNewLocation(55.7558, 37.6173)
		if loc.Latitude() != 55.7558 {
			t.Errorf("MustNewLocation() latitude = %v, want 55.7558", loc.Latitude())
		}
	})

	t.Run("invalid location panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewLocation() should panic on invalid location")
			}
		}()
		_ = MustNewLocation(91.0, 37.6173)
	})
}

