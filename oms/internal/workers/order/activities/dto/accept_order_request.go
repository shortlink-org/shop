package dto

import (
	"errors"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// ErrNoDeliveryInfo is returned when the order has no delivery info (nil).
// Caller should ensure order.HasDeliveryInfo() before calling AcceptOrderRequestFromOrder,
// or handle this error explicitly.
var ErrNoDeliveryInfo = errors.New("order has no delivery info")

// AcceptOrderRequestFromOrder builds ports.AcceptOrderRequest from an order that has delivery info.
// OrderID and CustomerID are taken from the aggregate (single source of truth).
// Returns ErrNoDeliveryInfo if order.GetDeliveryInfo() is nil, or an error from priority mapping.
func AcceptOrderRequestFromOrder(order *orderv1.OrderState) (ports.AcceptOrderRequest, error) {
	info := order.GetDeliveryInfo()
	if info == nil {
		return ports.AcceptOrderRequest{}, ErrNoDeliveryInfo
	}

	priorityDTO, err := domainPriorityToDTO(info.GetPriority())
	if err != nil {
		return ports.AcceptOrderRequest{}, err
	}

	pickup := info.GetPickupAddress()
	delivery := info.GetDeliveryAddress()
	period := info.GetDeliveryPeriod()
	pkg := info.GetPackageInfo()

	req := ports.AcceptOrderRequest{
		OrderID:    order.GetOrderID(),
		CustomerID: order.GetCustomerId(),
		PickupAddress: ports.DeliveryAddress{
			Street:     pickup.Street(),
			City:       pickup.City(),
			PostalCode: pickup.PostalCode(),
			Country:    pickup.Country(),
			Latitude:   pickup.Latitude(),
			Longitude:  pickup.Longitude(),
		},
		DeliveryAddress: ports.DeliveryAddress{
			Street:     delivery.Street(),
			City:       delivery.City(),
			PostalCode: delivery.PostalCode(),
			Country:    delivery.Country(),
			Latitude:   delivery.Latitude(),
			Longitude:  delivery.Longitude(),
		},
		DeliveryPeriod: ports.DeliveryPeriodDTO{
			StartTime: period.GetStartTime(),
			EndTime:   period.GetEndTime(),
		},
		PackageInfo: ports.PackageInfoDTO{
			WeightKg: pkg.GetWeightKg(),
		},
		Priority: priorityDTO,
	}

	if rc := info.GetRecipientContacts(); rc != nil {
		req.RecipientName = rc.GetName()
		req.RecipientPhone = rc.GetPhone()
		req.RecipientEmail = rc.GetEmail()
	}

	return req, nil
}
