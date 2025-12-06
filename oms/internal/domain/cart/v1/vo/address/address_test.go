package address

import (
	"strings"
	"testing"

	"github.com/shortlink-org/shop/oms/internal/domain/cart/v1/vo/location"
)

func TestNewAddress(t *testing.T) {
	tests := []struct {
		name      string
		street    string
		city      string
		postalCode string
		country   string
		wantErr   bool
		errType   error
	}{
		{
			name:      "valid address",
			street:    "123 Main St",
			city:      "Moscow",
			postalCode: "101000",
			country:   "Russia",
			wantErr:   false,
		},
		{
			name:      "valid address without postal code",
			street:    "456 Oak Ave",
			city:      "St. Petersburg",
			postalCode: "",
			country:   "Russia",
			wantErr:   false,
		},
		{
			name:      "empty street",
			street:    "",
			city:      "Moscow",
			postalCode: "101000",
			country:   "Russia",
			wantErr:   true,
			errType:   ErrAddressStreetEmpty,
		},
		{
			name:      "empty city",
			street:    "123 Main St",
			city:      "",
			postalCode: "101000",
			country:   "Russia",
			wantErr:   true,
			errType:   ErrAddressCityEmpty,
		},
		{
			name:      "empty country",
			street:    "123 Main St",
			city:      "Moscow",
			postalCode: "101000",
			country:   "",
			wantErr:   true,
			errType:   ErrAddressCountryEmpty,
		},
		{
			name:      "whitespace only street",
			street:    "   ",
			city:      "Moscow",
			postalCode: "101000",
			country:   "Russia",
			wantErr:   true,
			errType:   ErrAddressStreetEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewAddress(tt.street, tt.city, tt.postalCode, tt.country)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAddress() expected error but got none")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewAddress() error = %v, want %v", err, tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("NewAddress() unexpected error: %v", err)
					return
				}
				if addr.Street() != strings.TrimSpace(tt.street) {
					t.Errorf("NewAddress() street = %v, want %v", addr.Street(), tt.street)
				}
				if addr.City() != strings.TrimSpace(tt.city) {
					t.Errorf("NewAddress() city = %v, want %v", addr.City(), tt.city)
				}
				if addr.Country() != strings.TrimSpace(tt.country) {
					t.Errorf("NewAddress() country = %v, want %v", addr.Country(), tt.country)
				}
			}
		})
	}
}

func TestNewAddressWithLocation(t *testing.T) {
	tests := []struct {
		name      string
		street    string
		city      string
		postalCode string
		country   string
		location  location.Location
		wantErr   bool
	}{
		{
			name:      "valid address with location",
			street:    "123 Main St",
			city:      "Moscow",
			postalCode: "101000",
			country:   "Russia",
			location:  location.MustNewLocation(55.7558, 37.6173),
			wantErr:   false,
		},
		{
			name:      "valid address with zero location",
			street:    "123 Main St",
			city:      "Moscow",
			postalCode: "101000",
			country:   "Russia",
			location:  location.MustNewLocation(0, 0),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewAddressWithLocation(tt.street, tt.city, tt.postalCode, tt.country, tt.location)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAddressWithLocation() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("NewAddressWithLocation() unexpected error: %v", err)
					return
				}
				if addr.Location() != tt.location {
					t.Errorf("NewAddressWithLocation() location = %v, want %v", addr.Location(), tt.location)
				}
				if addr.Latitude() != tt.location.Latitude() {
					t.Errorf("NewAddressWithLocation() latitude = %v, want %v", addr.Latitude(), tt.location.Latitude())
				}
				if addr.Longitude() != tt.location.Longitude() {
					t.Errorf("NewAddressWithLocation() longitude = %v, want %v", addr.Longitude(), tt.location.Longitude())
				}
			}
		})
	}
}

func TestAddress_Getters(t *testing.T) {
	street := "123 Main St"
	city := "Moscow"
	postalCode := "101000"
	country := "Russia"

	addr, err := NewAddress(street, city, postalCode, country)
	if err != nil {
		t.Fatalf("NewAddress() error = %v", err)
	}

	if addr.Street() != street {
		t.Errorf("Address.Street() = %v, want %v", addr.Street(), street)
	}
	if addr.City() != city {
		t.Errorf("Address.City() = %v, want %v", addr.City(), city)
	}
	if addr.PostalCode() != postalCode {
		t.Errorf("Address.PostalCode() = %v, want %v", addr.PostalCode(), postalCode)
	}
	if addr.Country() != country {
		t.Errorf("Address.Country() = %v, want %v", addr.Country(), country)
	}
}

func TestAddress_HasCoordinates(t *testing.T) {
	tests := []struct {
		name     string
		location location.Location
		want     bool
	}{
		{
			name:     "has coordinates",
			location: location.MustNewLocation(55.7558, 37.6173),
			want:     true,
		},
		{
			name:     "no coordinates (zero)",
			location: location.MustNewLocation(0, 0),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewAddressWithLocation("123 Main St", "Moscow", "101000", "Russia", tt.location)
			if err != nil {
				t.Fatalf("NewAddressWithLocation() error = %v", err)
			}
			if addr.HasCoordinates() != tt.want {
				t.Errorf("Address.HasCoordinates() = %v, want %v", addr.HasCoordinates(), tt.want)
			}
		})
	}
}

func TestAddress_IsValid(t *testing.T) {
	validAddr, _ := NewAddress("123 Main St", "Moscow", "101000", "Russia")
	if !validAddr.IsValid() {
		t.Errorf("Address.IsValid() = false for valid address")
	}
}

func TestAddress_String(t *testing.T) {
	tests := []struct {
		name      string
		street    string
		city      string
		postalCode string
		country   string
		want      string
	}{
		{
			name:      "full address",
			street:    "123 Main St",
			city:      "Moscow",
			postalCode: "101000",
			country:   "Russia",
			want:      "123 Main St, 101000, Moscow, Russia",
		},
		{
			name:      "address without postal code",
			street:    "456 Oak Ave",
			city:      "St. Petersburg",
			postalCode: "",
			country:   "Russia",
			want:      "456 Oak Ave, St. Petersburg, Russia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewAddress(tt.street, tt.city, tt.postalCode, tt.country)
			if err != nil {
				t.Fatalf("NewAddress() error = %v", err)
			}
			got := addr.String()
			if got != tt.want {
				t.Errorf("Address.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_FullString(t *testing.T) {
	loc := location.MustNewLocation(55.7558, 37.6173)
	addr, err := NewAddressWithLocation("123 Main St", "Moscow", "101000", "Russia", loc)
	if err != nil {
		t.Fatalf("NewAddressWithLocation() error = %v", err)
	}

	fullStr := addr.FullString()
	if !strings.Contains(fullStr, "123 Main St") {
		t.Errorf("Address.FullString() should contain street")
	}
	if !strings.Contains(fullStr, "55.7558") {
		t.Errorf("Address.FullString() should contain latitude")
	}
	if !strings.Contains(fullStr, "37.6173") {
		t.Errorf("Address.FullString() should contain longitude")
	}
}

func TestAddress_Equality(t *testing.T) {
	addr1, _ := NewAddress("123 Main St", "Moscow", "101000", "Russia")
	addr2, _ := NewAddress("123 Main St", "Moscow", "101000", "Russia")
	addr3, _ := NewAddress("456 Oak Ave", "Moscow", "101000", "Russia")

	if addr1 != addr2 {
		t.Errorf("Addresses with same values should be equal")
	}
	if addr1 == addr3 {
		t.Errorf("Addresses with different values should not be equal")
	}
}

func TestAddress_TrimsWhitespace(t *testing.T) {
	addr, err := NewAddress("  123 Main St  ", "  Moscow  ", "  101000  ", "  Russia  ")
	if err != nil {
		t.Fatalf("NewAddress() error = %v", err)
	}

	if addr.Street() != "123 Main St" {
		t.Errorf("Address should trim whitespace from street, got %v", addr.Street())
	}
	if addr.City() != "Moscow" {
		t.Errorf("Address should trim whitespace from city, got %v", addr.City())
	}
	if addr.Country() != "Russia" {
		t.Errorf("Address should trim whitespace from country, got %v", addr.Country())
	}
}

