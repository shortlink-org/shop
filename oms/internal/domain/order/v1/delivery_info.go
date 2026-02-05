package v1

import (
	"github.com/google/uuid"

	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
)

// RecipientContacts contains contact details for the delivery recipient.
type RecipientContacts struct {
	name  string
	phone string
	email string
}

// NewRecipientContacts creates RecipientContacts (all fields optional).
func NewRecipientContacts(name, phone, email string) RecipientContacts {
	return RecipientContacts{name: name, phone: phone, email: email}
}

// GetName returns the recipient name.
func (r RecipientContacts) GetName() string { return r.name }

// GetPhone returns the recipient phone.
func (r RecipientContacts) GetPhone() string { return r.phone }

// GetEmail returns the recipient email.
func (r RecipientContacts) GetEmail() string { return r.email }

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
	// recipientContacts is optional contact details for the recipient
	recipientContacts *RecipientContacts
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
// RecipientContacts is optional (nil allowed).
func NewDeliveryInfo(
	pickupAddr address.Address,
	deliveryAddr address.Address,
	deliveryPeriod DeliveryPeriod,
	packageInfo PackageInfo,
	priority DeliveryPriority,
	recipientContacts *RecipientContacts,
) DeliveryInfo {
	return DeliveryInfo{
		pickupAddress:     pickupAddr,
		deliveryAddress:   deliveryAddr,
		deliveryPeriod:    deliveryPeriod,
		packageInfo:       packageInfo,
		priority:          priority,
		recipientContacts: recipientContacts,
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

// GetRecipientContacts returns the optional recipient contacts.
func (d DeliveryInfo) GetRecipientContacts() *RecipientContacts {
	return d.recipientContacts
}

// IsValid checks if the delivery info is valid.
func (d DeliveryInfo) IsValid() bool {
	return d.pickupAddress.IsValid() &&
		d.deliveryAddress.IsValid() &&
		d.deliveryPeriod.IsValid() &&
		d.packageInfo.IsValid()
}
