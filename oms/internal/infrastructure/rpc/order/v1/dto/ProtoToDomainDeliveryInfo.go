package dto

import (
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common/v1"
	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
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
	pkgInfo := orderDomain.NewPackageInfo(
		protoInfo.GetPackageInfo().GetWeightKg(),
		protoInfo.GetPackageInfo().GetDimensions(),
	)

	// Convert priority
	priority := protoPriorityToDomain(protoInfo.GetPriority())

	deliveryInfo := orderDomain.NewDeliveryInfo(
		pickupAddr,
		deliveryAddr,
		period,
		pkgInfo,
		priority,
	)

	return &deliveryInfo
}

// protoAddressToDomain converts proto DeliveryAddress to domain Address.
func protoAddressToDomain(protoAddr *commonv1.DeliveryAddress) address.Address {
	if protoAddr == nil {
		return address.Address{}
	}

	loc, _ := location.NewLocation(protoAddr.GetLatitude(), protoAddr.GetLongitude())
	addr, _ := address.NewAddressWithLocation(
		protoAddr.GetStreet(),
		protoAddr.GetCity(),
		protoAddr.GetPostalCode(),
		protoAddr.GetCountry(),
		loc,
	)

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
