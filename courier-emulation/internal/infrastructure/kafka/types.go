package kafka

import "time"

// defaultLocationAccuracy is the default accuracy radius in meters for Location in events.
const defaultLocationAccuracy = 10.0

// PickUpOrderEvent represents an order picked up event.
type PickUpOrderEvent struct {
	OrderID        string    `json:"order_id"`
	CourierID      string    `json:"courier_id"`
	PickupLocation Location  `json:"pickup_location"`
	PickedUpAt     time.Time `json:"picked_up_at"`
}

// DeliverOrderEvent represents an order delivered event.
type DeliverOrderEvent struct {
	OrderID         string             `json:"order_id"`
	CourierID       string             `json:"courier_id"`
	Status          DeliveryStatus     `json:"status"`
	Reason          NotDeliveredReason `json:"reason,omitempty"`
	CurrentLocation Location           `json:"current_location"`
	DeliveredAt     time.Time          `json:"delivered_at"`
}

// Location represents a geographic location in events.
// Timestamps are always UTC.
type Location struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy"`
	Timestamp time.Time `json:"timestamp"`
}

// DeliveryStatus is the outcome of a delivery attempt.
type DeliveryStatus string

const (
	DeliveryStatusDelivered    DeliveryStatus = "DELIVERED"
	DeliveryStatusNotDelivered DeliveryStatus = "NOT_DELIVERED"
)

// NotDeliveredReason is the reason for a failed delivery (NOT_DELIVERED).
type NotDeliveredReason string

const (
	ReasonCustomerNotAvailable NotDeliveredReason = "CUSTOMER_NOT_AVAILABLE"
	ReasonWrongAddress         NotDeliveredReason = "WRONG_ADDRESS"
	ReasonCustomerRefused      NotDeliveredReason = "CUSTOMER_REFUSED"
	ReasonAccessDenied         NotDeliveredReason = "ACCESS_DENIED"
	ReasonPackageDamaged       NotDeliveredReason = "PACKAGE_DAMAGED"
	ReasonOther                NotDeliveredReason = "OTHER"
)

// validNotDeliveredReasons is the whitelist for NOT_DELIVERED reason.
var validNotDeliveredReasons = map[NotDeliveredReason]struct{}{
	ReasonCustomerNotAvailable: {},
	ReasonWrongAddress:         {},
	ReasonCustomerRefused:      {},
	ReasonAccessDenied:         {},
	ReasonPackageDamaged:       {},
	ReasonOther:                {},
}
