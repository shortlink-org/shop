package dto

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	v2 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
)

func DomainToOrderState(in *v1.OrderState) *v2.OrderState {
	if in == nil {
		return nil
	}

	domainItems := in.GetItems()
	// Map Items
	items := make([]*v2.OrderItem, len(domainItems))
	for i, item := range domainItems {
		items[i] = &v2.OrderItem{
			Id:       item.GetGoodId().String(),
			Quantity: item.GetQuantity(),
			Price:    item.GetPrice().InexactFloat64(),
		}
	}

	var (
		packageID   string
		requestedAt *timestamppb.Timestamp
	)

	deliveryInfo := domainDeliveryInfoToProto(in.GetDeliveryInfo())
	if info := in.GetDeliveryInfo(); info != nil && info.GetPackageId() != nil {
		packageID = info.GetPackageId().String()
	}

	if ts := in.GetDeliveryRequestedAt(); ts != nil {
		requestedAt = timestamppb.New(*ts)
	}

	return &v2.OrderState{
		Id:             in.GetOrderID().String(),
		CustomerId:     in.GetCustomerId().String(),
		Items:          items,
		Status:         in.GetStatus(),
		DeliveryInfo:   deliveryInfo,
		DeliveryStatus: in.GetDeliveryStatus(),
		PackageId:      packageID,
		RequestedAt:    requestedAt,
	}
}

func domainDeliveryInfoToProto(info *v1.DeliveryInfo) *commonv1.DeliveryInfo {
	if info == nil {
		return nil
	}

	return &commonv1.DeliveryInfo{
		PickupAddress:     domainAddressToProto(info.GetPickupAddress()),
		DeliveryAddress:   domainAddressToProto(info.GetDeliveryAddress()),
		DeliveryPeriod:    domainDeliveryPeriodToProto(info.GetDeliveryPeriod()),
		PackageInfo:       domainPackageInfoToProto(info.GetPackageInfo()),
		Priority:          domainPriorityToProto(info.GetPriority()),
		RecipientContacts: domainRecipientContactsToProto(info.GetRecipientContacts()),
	}
}

func domainAddressToProto(addr interface {
	Street() string
	City() string
	PostalCode() string
	Country() string
	Latitude() float64
	Longitude() float64
}) *commonv1.DeliveryAddress {
	return &commonv1.DeliveryAddress{
		Street:     addr.Street(),
		City:       addr.City(),
		PostalCode: addr.PostalCode(),
		Country:    addr.Country(),
		Latitude:   addr.Latitude(),
		Longitude:  addr.Longitude(),
	}
}

func domainDeliveryPeriodToProto(period v1.DeliveryPeriod) *commonv1.DeliveryPeriod {
	return &commonv1.DeliveryPeriod{
		StartTime: timestamppb.New(period.GetStartTime()),
		EndTime:   timestamppb.New(period.GetEndTime()),
	}
}

func domainPackageInfoToProto(info v1.PackageInfo) *commonv1.PackageInfo {
	return &commonv1.PackageInfo{
		WeightKg: info.GetWeightKg(),
	}
}

func domainPriorityToProto(priority v1.DeliveryPriority) commonv1.DeliveryPriority {
	switch priority {
	case v1.DeliveryPriorityNormal:
		return commonv1.DeliveryPriority_DELIVERY_PRIORITY_NORMAL
	case v1.DeliveryPriorityUrgent:
		return commonv1.DeliveryPriority_DELIVERY_PRIORITY_URGENT
	default:
		return commonv1.DeliveryPriority_DELIVERY_PRIORITY_UNSPECIFIED
	}
}

func domainRecipientContactsToProto(contacts *v1.RecipientContacts) *commonv1.RecipientContacts {
	if contacts == nil {
		return nil
	}

	return &commonv1.RecipientContacts{
		RecipientName:  contacts.GetName(),
		RecipientPhone: contacts.GetPhone(),
		RecipientEmail: contacts.GetEmail(),
	}
}
