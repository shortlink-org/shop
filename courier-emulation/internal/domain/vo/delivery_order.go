package vo

import (
	"time"
)

// DeliveryOrder represents an order assigned to a courier for delivery.
type DeliveryOrder struct {
	orderID          string
	packageID        string
	pickupLocation   Location
	deliveryLocation Location
	assignedAt       time.Time
}

// NewDeliveryOrder creates a new DeliveryOrder.
func NewDeliveryOrder(
	orderID string,
	packageID string,
	pickupLocation Location,
	deliveryLocation Location,
	assignedAt time.Time,
) DeliveryOrder {
	return DeliveryOrder{
		orderID:          orderID,
		packageID:        packageID,
		pickupLocation:   pickupLocation,
		deliveryLocation: deliveryLocation,
		assignedAt:       assignedAt,
	}
}

// OrderID returns the order ID.
func (o DeliveryOrder) OrderID() string {
	return o.orderID
}

// PackageID returns the package ID.
func (o DeliveryOrder) PackageID() string {
	return o.packageID
}

// PickupLocation returns the pickup location.
func (o DeliveryOrder) PickupLocation() Location {
	return o.pickupLocation
}

// DeliveryLocation returns the delivery location.
func (o DeliveryOrder) DeliveryLocation() Location {
	return o.deliveryLocation
}

// AssignedAt returns the time the order was assigned.
func (o DeliveryOrder) AssignedAt() time.Time {
	return o.assignedAt
}

// DistanceToPickup calculates the distance from a location to the pickup point.
func (o DeliveryOrder) DistanceToPickup(from Location) float64 {
	return from.DistanceTo(o.pickupLocation)
}

// DistanceToDelivery calculates the distance from a location to the delivery point.
func (o DeliveryOrder) DistanceToDelivery(from Location) float64 {
	return from.DistanceTo(o.deliveryLocation)
}

// TotalDistance calculates the total distance from pickup to delivery.
func (o DeliveryOrder) TotalDistance() float64 {
	return o.pickupLocation.DistanceTo(o.deliveryLocation)
}
