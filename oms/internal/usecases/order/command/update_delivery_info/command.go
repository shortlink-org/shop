package update_delivery_info

import (
	"github.com/google/uuid"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Command represents a command to update order delivery information.
type Command struct {
	OrderID      uuid.UUID
	DeliveryInfo orderv1.DeliveryInfo
}

// NewCommand creates a new UpdateDeliveryInfo command.
func NewCommand(orderID uuid.UUID, deliveryInfo orderv1.DeliveryInfo) Command {
	return Command{
		OrderID:      orderID,
		DeliveryInfo: deliveryInfo,
	}
}
