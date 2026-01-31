package add_items

import (
	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// Command represents a command to add multiple items to a cart.
type Command struct {
	CustomerID uuid.UUID
	Items      []itemv1.Item
}

// NewCommand creates a new AddItems command.
func NewCommand(customerID uuid.UUID, items []itemv1.Item) Command {
	return Command{
		CustomerID: customerID,
		Items:      items,
	}
}
