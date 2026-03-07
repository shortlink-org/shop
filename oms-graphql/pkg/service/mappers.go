package service

import (
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	cartmodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/cart/v1/model/v1"
	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func mapCartState(state *cartmodel.CartState) *servicepb.CartState {
	if state == nil {
		return nil
	}

	items := make([]*servicepb.CartItem, 0, len(state.GetItems()))
	for _, item := range state.GetItems() {
		items = append(items, &servicepb.CartItem{
			GoodId:   wrapperspb.String(item.GetGoodId()),
			Quantity: wrapperspb.Int32(item.GetQuantity()),
		})
	}

	return &servicepb.CartState{
		CartId: wrapperspb.String(state.GetCartId()),
		Items: &servicepb.ListOfCartItem{
			List: &servicepb.ListOfCartItem_List{
				Items: items,
			},
		},
	}
}

func mapCartItemInputs(input *servicepb.ItemRequest) []*cartmodel.CartItem {
	if input == nil {
		return nil
	}

	items := make([]*cartmodel.CartItem, 0, len(input.GetItems()))
	for _, item := range input.GetItems() {
		items = append(items, &cartmodel.CartItem{
			GoodId:   item.GetGoodId(),
			Quantity: item.GetQuantity(),
		})
	}

	return items
}

func mapCreateOrderRequest(userID string, input *servicepb.CreateOrderInput) (*ordermodel.CreateRequest, error) {
	if input == nil {
		return nil, connectInvalidArgument("input is required")
	}

	orderID := uuid.NewString()
	deliveryInfo, err := mapDeliveryInfoInput(input.GetDeliveryInfo())
	if err != nil {
		return nil, err
	}

	items := make([]*ordermodel.OrderItem, 0, len(input.GetItems()))
	for _, item := range input.GetItems() {
		items = append(items, &ordermodel.OrderItem{
			Id:       item.GetId(),
			Quantity: item.GetQuantity(),
			Price:    item.GetPrice(),
		})
	}

	return &ordermodel.CreateRequest{
		Order: &ordermodel.OrderState{
			Id:         orderID,
			CustomerId: userID,
			Items:      items,
		},
		DeliveryInfo: deliveryInfo,
	}, nil
}

func mapOrderState(order *ordermodel.OrderState) *servicepb.OrderState {
	if order == nil {
		return nil
	}

	items := make([]*servicepb.OrderItem, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, &servicepb.OrderItem{
			Id:       wrapperspb.String(item.GetId()),
			Quantity: wrapperspb.Int32(item.GetQuantity()),
			Price:    wrapperspb.Double(item.GetPrice()),
		})
	}

	return &servicepb.OrderState{
		Id: wrapperspb.String(order.GetId()),
		Items: &servicepb.ListOfOrderItem{
			List: &servicepb.ListOfOrderItem_List{
				Items: items,
			},
		},
		Status:       wrapperspb.String(order.GetStatus().String()),
		DeliveryInfo: mapDeliveryInfo(order.GetDeliveryInfo()),
	}
}

func mapDeliveryInfo(info *commonpb.DeliveryInfo) *servicepb.DeliveryInfo {
	if info == nil {
		return nil
	}

	return &servicepb.DeliveryInfo{
		PickupAddress:     mapDeliveryAddress(info.GetPickupAddress()),
		DeliveryAddress:   mapDeliveryAddress(info.GetDeliveryAddress()),
		DeliveryPeriod:    mapDeliveryPeriod(info.GetDeliveryPeriod()),
		PackageInfo:       mapPackageInfo(info.GetPackageInfo()),
		Priority:          wrapperspb.String(info.GetPriority().String()),
		RecipientContacts: mapRecipientContacts(info.GetRecipientContacts()),
	}
}

func mapDeliveryAddress(address *commonpb.DeliveryAddress) *servicepb.DeliveryAddress {
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

func mapDeliveryPeriod(period *commonpb.DeliveryPeriod) *servicepb.DeliveryPeriod {
	if period == nil {
		return nil
	}

	return &servicepb.DeliveryPeriod{
		StartTime: wrapperspb.String(period.GetStartTime().AsTime().Format(time.RFC3339)),
		EndTime:   wrapperspb.String(period.GetEndTime().AsTime().Format(time.RFC3339)),
	}
}

func mapPackageInfo(info *commonpb.PackageInfo) *servicepb.PackageInfo {
	if info == nil {
		return nil
	}

	return &servicepb.PackageInfo{
		WeightKg: wrapperspb.Double(info.GetWeightKg()),
	}
}

func mapRecipientContacts(contacts *commonpb.RecipientContacts) *servicepb.RecipientContacts {
	if contacts == nil {
		return nil
	}

	return &servicepb.RecipientContacts{
		RecipientName:  wrapperspb.String(contacts.GetRecipientName()),
		RecipientPhone: wrapperspb.String(contacts.GetRecipientPhone()),
		RecipientEmail: wrapperspb.String(contacts.GetRecipientEmail()),
	}
}

func mapDeliveryInfoInput(input *servicepb.DeliveryInfoInput) (*commonpb.DeliveryInfo, error) {
	if input == nil {
		return nil, nil
	}

	startTime, err := parseTimestamp(input.GetDeliveryPeriod().GetStartTime())
	if err != nil {
		return nil, err
	}

	endTime, err := parseTimestamp(input.GetDeliveryPeriod().GetEndTime())
	if err != nil {
		return nil, err
	}

	return &commonpb.DeliveryInfo{
		PickupAddress:   mapDeliveryAddressInput(input.GetPickupAddress()),
		DeliveryAddress: mapDeliveryAddressInput(input.GetDeliveryAddress()),
		DeliveryPeriod: &commonpb.DeliveryPeriod{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
		PackageInfo: &commonpb.PackageInfo{
			WeightKg: input.GetPackageInfo().GetWeightKg(),
		},
		Priority:          parsePriority(input.GetPriority()),
		RecipientContacts: mapRecipientContactsInput(input.GetRecipientContacts()),
	}, nil
}

func mapDeliveryAddressInput(input *servicepb.DeliveryAddressInput) *commonpb.DeliveryAddress {
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

func mapRecipientContactsInput(input *servicepb.RecipientContactsInput) *commonpb.RecipientContacts {
	if input == nil {
		return nil
	}

	contacts := &commonpb.RecipientContacts{}
	if input.GetRecipientName() != nil {
		contacts.RecipientName = input.GetRecipientName().GetValue()
	}
	if input.GetRecipientPhone() != nil {
		contacts.RecipientPhone = input.GetRecipientPhone().GetValue()
	}
	if input.GetRecipientEmail() != nil {
		contacts.RecipientEmail = input.GetRecipientEmail().GetValue()
	}

	if contacts.RecipientName == "" && contacts.RecipientPhone == "" && contacts.RecipientEmail == "" {
		return nil
	}

	return contacts
}

func parseTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, nil
	}

	return time.Time{}, connectInvalidArgument("deliveryPeriod values must be RFC3339 timestamps")
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

func connectInvalidArgument(message string) error {
	return connectError(connect.CodeInvalidArgument, message)
}

func connectError(code connect.Code, message string) error {
	return connect.NewError(code, errors.New(message))
}
