package create

import (
	"github.com/google/uuid"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Command represents a command to create an order.
type Command struct {
	OrderID      uuid.UUID
	CustomerID   uuid.UUID
	Items        orderv1.Items
	DeliveryInfo *orderv1.DeliveryInfo
}

// NewCommand creates a new CreateOrder command.
func NewCommand(orderID, customerID uuid.UUID, items orderv1.Items, deliveryInfo *orderv1.DeliveryInfo) Command {
	return Command{
		OrderID:      orderID,
		CustomerID:   customerID,
		Items:        items,
		DeliveryInfo: deliveryInfo,
	}
}
