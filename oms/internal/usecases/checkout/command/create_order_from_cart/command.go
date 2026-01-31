package create_order_from_cart

import (
	"github.com/google/uuid"

	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Command represents a command to create an order from a cart.
type Command struct {
	CustomerID   uuid.UUID
	DeliveryInfo *orderDomain.DeliveryInfo
}

// NewCommand creates a new CreateOrderFromCart command.
func NewCommand(customerID uuid.UUID, deliveryInfo *orderDomain.DeliveryInfo) Command {
	return Command{
		CustomerID:   customerID,
		DeliveryInfo: deliveryInfo,
	}
}
