package kafka

import (
	"fmt"
	"time"

	"github.com/shortlink-org/shortlink/boundaries/shop/courier-emulation/internal/domain/vo"
)

// NewPickUpOrderEvent creates an order picked up event from domain objects.
func NewPickUpOrderEvent(courierID string, order vo.DeliveryOrder, location vo.Location) PickUpOrderEvent {
	now := time.Now().UTC()
	return PickUpOrderEvent{
		OrderID:   order.OrderID(),
		CourierID: courierID,
		PickupLocation: Location{
			Latitude:  location.Latitude(),
			Longitude: location.Longitude(),
			Accuracy:  defaultLocationAccuracy,
			Timestamp: now,
		},
		PickedUpAt: now,
	}
}

// NewDeliverOrderEvent creates an order delivered event from domain objects.
// Validates: when delivered is true, reason must be empty; when false, reason must be from whitelist (or OTHER).
func NewDeliverOrderEvent(
	courierID string,
	order vo.DeliveryOrder,
	location vo.Location,
	delivered bool,
	reason NotDeliveredReason,
) (DeliverOrderEvent, error) {
	if delivered && reason != "" {
		return DeliverOrderEvent{}, fmt.Errorf("%w: got=%q", ErrReasonMustBeEmpty, reason)
	}
	if !delivered {
		if reason == "" {
			return DeliverOrderEvent{}, fmt.Errorf("%w", ErrReasonRequired)
		}
		if _, ok := validNotDeliveredReasons[reason]; !ok {
			return DeliverOrderEvent{}, fmt.Errorf("%w: got=%q", ErrInvalidReason, reason)
		}
	}

	status := DeliveryStatusDelivered
	if !delivered {
		status = DeliveryStatusNotDelivered
	}

	now := time.Now().UTC()
	return DeliverOrderEvent{
		OrderID:   order.OrderID(),
		CourierID: courierID,
		Status:    status,
		Reason:    reason,
		CurrentLocation: Location{
			Latitude:  location.Latitude(),
			Longitude: location.Longitude(),
			Accuracy:  defaultLocationAccuracy,
			Timestamp: now,
		},
		DeliveredAt: now,
	}, nil
}
