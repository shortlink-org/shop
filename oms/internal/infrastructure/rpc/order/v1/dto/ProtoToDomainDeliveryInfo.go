package dto

import (
	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/address"
	"github.com/shortlink-org/shop/oms/internal/domain/order/v1/vo/location"
)

// ProtoDeliveryInfoToDomain converts proto DeliveryInfo to domain DeliveryInfo.
// Returns nil if protoInfo is nil.
func ProtoDeliveryInfoToDomain(protoInfo *commonv1.DeliveryInfo) *orderDomain.DeliveryInfo {
	if protoInfo == nil {
		return nil
	}

	// Convert pickup address
	pickupAddr := protoAddressToDomain(protoInfo.GetPickupAddress())

	// Convert delivery address
	deliveryAddr := protoAddressToDomain(protoInfo.GetDeliveryAddress())

	// Convert delivery period
	period := orderDomain.NewDeliveryPeriod(
		protoInfo.GetDeliveryPeriod().GetStartTime().AsTime(),
		protoInfo.GetDeliveryPeriod().GetEndTime().AsTime(),
	)

	// Convert package info
	pkgInfo := orderDomain.NewPackageInfo(protoInfo.GetPackageInfo().GetWeightKg())

	// Convert priority
	priority := protoPriorityToDomain(protoInfo.GetPriority())

	// Convert optional recipient contacts
	var recipientContacts *orderDomain.RecipientContacts

	if rc := protoInfo.GetRecipientContacts(); rc != nil {
		c := orderDomain.NewRecipientContacts(
			rc.GetRecipientName(),
			rc.GetRecipientPhone(),
			rc.GetRecipientEmail(),
		)
		recipientContacts = &c
	}

	deliveryInfo := orderDomain.NewDeliveryInfo(
		pickupAddr,
		deliveryAddr,
		period,
		pkgInfo,
		priority,
		recipientContacts,
	)

	return &deliveryInfo
}

// protoAddressToDomain converts proto DeliveryAddress to domain Address.
func protoAddressToDomain(protoAddr *commonv1.DeliveryAddress) address.Address {
	if protoAddr == nil {
		return address.Address{}
	}

	loc, err := location.NewLocation(protoAddr.GetLatitude(), protoAddr.GetLongitude())
	if err != nil {
		return address.Address{}
	}

	addr, err := address.NewAddressWithLocation(
		protoAddr.GetStreet(),
		protoAddr.GetCity(),
		protoAddr.GetPostalCode(),
		protoAddr.GetCountry(),
		loc,
	)
	if err != nil {
		return address.Address{}
	}

	return addr
}

// protoPriorityToDomain converts proto DeliveryPriority to domain DeliveryPriority.
func protoPriorityToDomain(protoPriority commonv1.DeliveryPriority) orderDomain.DeliveryPriority {
	switch protoPriority {
	case commonv1.DeliveryPriority_DELIVERY_PRIORITY_NORMAL:
		return orderDomain.DeliveryPriorityNormal
	case commonv1.DeliveryPriority_DELIVERY_PRIORITY_URGENT:
		return orderDomain.DeliveryPriorityUrgent
	default:
		return orderDomain.DeliveryPriorityUnspecified
	}
}
