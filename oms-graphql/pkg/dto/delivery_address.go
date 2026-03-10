package dto

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

// DeliveryAddressToService maps OMS delivery address to Connect response.
func DeliveryAddressToService(address *commonpb.DeliveryAddress) *servicepb.DeliveryAddress {
	if address == nil {
		return nil
	}

	return &servicepb.DeliveryAddress{
		Street:     wrapperspb.String(address.GetStreet()),
		City:       wrapperspb.String(address.GetCity()),
		PostalCode: wrapperspb.String(address.GetPostalCode()),
		Country:    wrapperspb.String(address.GetCountry()),
		Latitude:   wrapperspb.Double(address.GetLatitude()),
		Longitude:  wrapperspb.Double(address.GetLongitude()),
	}
}

// DeliveryAddressFromInput maps Connect delivery address input to OMS proto.
func DeliveryAddressFromInput(input *servicepb.DeliveryAddressInput) *commonpb.DeliveryAddress {
	if input == nil {
		return nil
	}

	address := &commonpb.DeliveryAddress{
		Street:  input.GetStreet(),
		City:    input.GetCity(),
		Country: input.GetCountry(),
	}
	if input.GetPostalCode() != nil {
		address.PostalCode = input.GetPostalCode().GetValue()
	}

	if input.GetLatitude() != nil {
		address.Latitude = input.GetLatitude().GetValue()
	}

	if input.GetLongitude() != nil {
		address.Longitude = input.GetLongitude().GetValue()
	}

	return address
}
