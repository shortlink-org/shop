package address

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/location"
)

// Address validation errors
var (
	ErrAddressStreetEmpty  = errors.New("address street cannot be empty")
	ErrAddressCityEmpty    = errors.New("address city cannot be empty")
	ErrAddressCountryEmpty = errors.New("address country cannot be empty")
)

// Address represents a physical address with coordinates as a value object.
// A value object is immutable and defined by its attributes.
// Two addresses are considered equal if all their fields match.
type Address struct {
	// street is the street address
	street string
	// city is the city name
	city string
	// postalCode is the postal/zip code (optional)
	postalCode string
	// country is the country name
	country string
	// location is the GPS location (optional, zero value if not set)
	location location.Location
}

// NewAddress creates a new Address value object with validation.
//
// Args:
//   - street: Street address (required, cannot be empty)
//   - city: City name (required, cannot be empty)
//   - postalCode: Postal/zip code (optional, can be empty)
//   - country: Country name (required, cannot be empty)
//
// Returns:
//   - Address: The validated address value object
//   - error: Error if address is invalid
//
// Example:
//
//	addr, err := vo.NewAddress("123 Main St", "Moscow", "101000", "Russia")
//	if err != nil {
//	    return err
//	}
func NewAddress(street, city, postalCode, country string) (Address, error) {
	street = strings.TrimSpace(street)
	if street == "" {
		return Address{}, ErrAddressStreetEmpty
	}

	city = strings.TrimSpace(city)
	if city == "" {
		return Address{}, ErrAddressCityEmpty
	}

	country = strings.TrimSpace(country)
	if country == "" {
		return Address{}, ErrAddressCountryEmpty
	}

	return Address{
		street:     street,
		city:       city,
		postalCode: strings.TrimSpace(postalCode),
		country:    country,
		location:   location.Location{}, // zero location
	}, nil
}

// NewAddressWithLocation creates a new Address value object with location.
//
// Args:
//   - street: Street address (required, cannot be empty)
//   - city: City name (required, cannot be empty)
//   - postalCode: Postal/zip code (optional, can be empty)
//   - country: Country name (required, cannot be empty)
//   - location: GPS location value object
//
// Returns:
//   - Address: The validated address value object
//   - error: Error if address is invalid
func NewAddressWithLocation(street, city, postalCode, country string, loc location.Location) (Address, error) {
	addr, err := NewAddress(street, city, postalCode, country)
	if err != nil {
		return Address{}, err
	}

	addr.location = loc

	return addr, nil
}

// Street returns the street address.
func (a Address) Street() string {
	return a.street
}

// City returns the city name.
func (a Address) City() string {
	return a.city
}

// PostalCode returns the postal/zip code.
func (a Address) PostalCode() string {
	return a.postalCode
}

// Country returns the country name.
func (a Address) Country() string {
	return a.country
}

// Location returns the GPS location.
func (a Address) Location() location.Location {
	return a.location
}

// Latitude returns the latitude coordinate (0 if location is not set).
func (a Address) Latitude() float64 {
	return a.location.Latitude()
}

// Longitude returns the longitude coordinate (0 if location is not set).
func (a Address) Longitude() float64 {
	return a.location.Longitude()
}

// HasCoordinates checks if the address has a valid location set.
func (a Address) HasCoordinates() bool {
	return !a.location.IsZero()
}

// IsValid checks if the address is valid.
func (a Address) IsValid() bool {
	return a.street != "" && a.city != "" && a.country != ""
}

// String returns a formatted string representation of the address.
func (a Address) String() string {
	var parts []string
	if a.street != "" {
		parts = append(parts, a.street)
	}
	if a.postalCode != "" {
		parts = append(parts, a.postalCode)
	}
	if a.city != "" {
		parts = append(parts, a.city)
	}
	if a.country != "" {
		parts = append(parts, a.country)
	}
	return strings.Join(parts, ", ")
}

// FullString returns a detailed string representation including coordinates if available.
func (a Address) FullString() string {
	base := a.String()
	if a.HasCoordinates() {
		return fmt.Sprintf("%s %s", base, a.location.String())
	}
	return base
}
