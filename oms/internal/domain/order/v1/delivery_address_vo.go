package v1

// NewDeliveryAddress creates a new DeliveryAddress from proto message.
// Helper function to create DeliveryAddress instances.
// Coordinates (latitude, longitude) are optional and set to 0 by default.
func NewDeliveryAddress(street, city, postalCode, country string) *DeliveryAddress {
	return &DeliveryAddress{
		Street:     street,
		City:       city,
		PostalCode: postalCode,
		Country:    country,
		// Latitude and Longitude are set to 0 by default
		Latitude:  0,
		Longitude: 0,
	}
}

// IsDeliveryAddressValid checks if the address has required fields.
func IsDeliveryAddressValid(addr *DeliveryAddress) bool {
	if addr == nil {
		return false
	}
	return addr.Street != "" && addr.City != "" && addr.Country != ""
}

