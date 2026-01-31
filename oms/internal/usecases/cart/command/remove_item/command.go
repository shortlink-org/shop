package remove_item

import (
	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// Command represents a command to remove an item from a cart.
type Command struct {
	CustomerID uuid.UUID
	Item       itemv1.Item
}

// NewCommand creates a new RemoveItem command.
func NewCommand(customerID uuid.UUID, item itemv1.Item) Command {
	return Command{
		CustomerID: customerID,
		Item:       item,
	}
}
