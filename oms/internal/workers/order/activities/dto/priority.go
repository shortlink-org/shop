package dto

import (
	"errors"
	"fmt"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// ErrUnsupportedDeliveryPriority is returned when the domain priority value
// cannot be mapped to the port DTO (e.g. unknown enum value).
var ErrUnsupportedDeliveryPriority = errors.New("unsupported delivery priority")

// domainPriorityToDTO maps domain DeliveryPriority to the port DeliveryPriorityDTO.
// Returns an error for unknown values so that enum contract drift is detected.
func domainPriorityToDTO(p orderv1.DeliveryPriority) (ports.DeliveryPriorityDTO, error) {
	switch p {
	case orderv1.DeliveryPriorityUnspecified:
		return ports.DeliveryPriorityUnspecified, nil
	case orderv1.DeliveryPriorityNormal:
		return ports.DeliveryPriorityNormal, nil
	case orderv1.DeliveryPriorityUrgent:
		return ports.DeliveryPriorityUrgent, nil
	default:
		return 0, fmt.Errorf("%w: %d", ErrUnsupportedDeliveryPriority, p)
	}
}
