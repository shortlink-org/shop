package dto

import (
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

// DeliveryInfoToService maps OMS delivery info to Connect/GraphQL response.
func DeliveryInfoToService(info *commonpb.DeliveryInfo) *servicepb.DeliveryInfo {
	if info == nil {
		return nil
	}

	return &servicepb.DeliveryInfo{
		PickupAddress:     DeliveryAddressToService(info.GetPickupAddress()),
		DeliveryAddress:   DeliveryAddressToService(info.GetDeliveryAddress()),
		DeliveryPeriod:    DeliveryPeriodToService(info.GetDeliveryPeriod()),
		PackageInfo:       PackageInfoToService(info.GetPackageInfo()),
		Priority:          wrapperspb.String(info.GetPriority().String()),
		RecipientContacts: RecipientContactsToService(info.GetRecipientContacts()),
	}
}

// DeliveryInfoFromInput maps Connect delivery info input to OMS proto.
func DeliveryInfoFromInput(input *servicepb.DeliveryInfoInput) (*commonpb.DeliveryInfo, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // optional input, no error
	}

	var deliveryPeriod *commonpb.DeliveryPeriod

	if p := input.GetDeliveryPeriod(); p != nil {
		startTime, err := ParseTimestamp(p.GetStartTime())
		if err != nil {
			return nil, err
		}

		endTime, err := ParseTimestamp(p.GetEndTime())
		if err != nil {
			return nil, err
		}

		deliveryPeriod = &commonpb.DeliveryPeriod{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		}
	}

	var packageInfo *commonpb.PackageInfo
	if pi := input.GetPackageInfo(); pi != nil {
		packageInfo = &commonpb.PackageInfo{WeightKg: pi.GetWeightKg()}
	}

	return &commonpb.DeliveryInfo{
		PickupAddress:     DeliveryAddressFromInput(input.GetPickupAddress()),
		DeliveryAddress:   DeliveryAddressFromInput(input.GetDeliveryAddress()),
		DeliveryPeriod:    deliveryPeriod,
		PackageInfo:       packageInfo,
		Priority:          parsePriority(input.GetPriority()),
		RecipientContacts: RecipientContactsFromInput(input.GetRecipientContacts()),
	}, nil
}

func parsePriority(priority *wrapperspb.StringValue) commonpb.DeliveryPriority {
	if priority == nil {
		return commonpb.DeliveryPriority_DELIVERY_PRIORITY_UNSPECIFIED
	}

	switch strings.ToUpper(strings.TrimSpace(priority.GetValue())) {
	case "", "DELIVERY_PRIORITY_UNSPECIFIED", "UNSPECIFIED":
		return commonpb.DeliveryPriority_DELIVERY_PRIORITY_UNSPECIFIED
	case "DELIVERY_PRIORITY_NORMAL", "NORMAL":
		return commonpb.DeliveryPriority_DELIVERY_PRIORITY_NORMAL
	case "DELIVERY_PRIORITY_URGENT", "URGENT":
		return commonpb.DeliveryPriority_DELIVERY_PRIORITY_URGENT
	default:
		return commonpb.DeliveryPriority_DELIVERY_PRIORITY_UNSPECIFIED
	}
}
