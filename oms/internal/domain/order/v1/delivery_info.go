package v1

import (
	"github.com/google/uuid"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
)

// DeliveryInfo represents delivery information for an order.
// This is a value object that aggregates delivery-related data.
type DeliveryInfo struct {
	// pickupAddress is the address where the package will be picked up
	pickupAddress address.Address
	// deliveryAddress is the address where the package should be delivered
	deliveryAddress address.Address
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

// String returns the string representation of the delivery priority.
func (p DeliveryPriority) String() string {
	switch p {
	case DeliveryPriorityNormal:
		return "NORMAL"
	case DeliveryPriorityUrgent:
		return "URGENT"
	default:
		return "UNSPECIFIED"
	}
}

// DeliveryPriorityFromString converts a string to DeliveryPriority.
func DeliveryPriorityFromString(s string) DeliveryPriority {
	switch s {
	case "NORMAL":
		return DeliveryPriorityNormal
	case "URGENT":
		return DeliveryPriorityUrgent
	default:
		return DeliveryPriorityUnspecified
	}
}

// NewDeliveryInfo creates a new DeliveryInfo value object.
func NewDeliveryInfo(
	pickupAddr address.Address,
	deliveryAddr address.Address,
	deliveryPeriod DeliveryPeriod,
	packageInfo PackageInfo,
	priority DeliveryPriority,
) DeliveryInfo {
	return DeliveryInfo{
		pickupAddress:   pickupAddr,
		deliveryAddress: deliveryAddr,
		deliveryPeriod:  deliveryPeriod,
		packageInfo:     packageInfo,
		priority:        priority,
	}
}

// GetPickupAddress returns the pickup address.
func (d DeliveryInfo) GetPickupAddress() address.Address {
	return d.pickupAddress
}

// GetDeliveryAddress returns the delivery address.
func (d DeliveryInfo) GetDeliveryAddress() address.Address {
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
	return d.pickupAddress.IsValid() &&
		d.deliveryAddress.IsValid() &&
		d.deliveryPeriod.IsValid() &&
		d.packageInfo.IsValid()
}
