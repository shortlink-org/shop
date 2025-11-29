package v1

import "github.com/google/uuid"

// DeliveryInfo represents delivery information for an order.
// This is a value object that aggregates delivery-related data.
type DeliveryInfo struct {
	// pickupAddress is the address where the package will be picked up
	pickupAddress *DeliveryAddress
	// deliveryAddress is the address where the package should be delivered
	deliveryAddress *DeliveryAddress
	// deliveryPeriod is the desired delivery time window
	deliveryPeriod DeliveryPeriod
	// packageInfo contains physical characteristics of the package
	packageInfo PackageInfo
	// packageId is the ID assigned by the delivery service (set after order is sent to delivery)
	packageId *uuid.UUID
	// priority indicates the delivery priority level
	priority DeliveryPriority
}

// DeliveryPriority represents delivery priority level.
type DeliveryPriority int32

const (
	DeliveryPriorityUnspecified DeliveryPriority = 0
	DeliveryPriorityNormal      DeliveryPriority = 1
	DeliveryPriorityUrgent      DeliveryPriority = 2
)

// NewDeliveryInfo creates a new DeliveryInfo value object.
func NewDeliveryInfo(
	pickupAddress *DeliveryAddress,
	deliveryAddress *DeliveryAddress,
	deliveryPeriod DeliveryPeriod,
	packageInfo PackageInfo,
	priority DeliveryPriority,
) DeliveryInfo {
	return DeliveryInfo{
		pickupAddress:   pickupAddress,
		deliveryAddress: deliveryAddress,
		deliveryPeriod:  deliveryPeriod,
		packageInfo:     packageInfo,
		priority:        priority,
	}
}

// GetPickupAddress returns the pickup address.
func (d DeliveryInfo) GetPickupAddress() *DeliveryAddress {
	return d.pickupAddress
}

// GetDeliveryAddress returns the delivery address.
func (d DeliveryInfo) GetDeliveryAddress() *DeliveryAddress {
	return d.deliveryAddress
}

// GetDeliveryPeriod returns the delivery period.
func (d DeliveryInfo) GetDeliveryPeriod() DeliveryPeriod {
	return d.deliveryPeriod
}

// GetPackageInfo returns the package info.
func (d DeliveryInfo) GetPackageInfo() PackageInfo {
	return d.packageInfo
}

// GetPackageId returns the package ID if set.
func (d DeliveryInfo) GetPackageId() *uuid.UUID {
	return d.packageId
}

// SetPackageId sets the package ID (called after delivery service accepts the order).
func (d *DeliveryInfo) SetPackageId(packageId uuid.UUID) {
	d.packageId = &packageId
}

// GetPriority returns the delivery priority.
func (d DeliveryInfo) GetPriority() DeliveryPriority {
	return d.priority
}

// IsValid checks if the delivery info is valid.
func (d DeliveryInfo) IsValid() bool {
	return d.pickupAddress != nil && IsDeliveryAddressValid(d.pickupAddress) &&
		d.deliveryAddress != nil && IsDeliveryAddressValid(d.deliveryAddress) &&
		d.deliveryPeriod.IsValid() &&
		d.packageInfo.IsValid()
}
