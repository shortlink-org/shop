package cancel

import (
	"github.com/google/uuid"
)

// Command represents a command to cancel an order.
type Command struct {
	OrderID uuid.UUID
}

// NewCommand creates a new CancelOrder command.
func NewCommand(orderID uuid.UUID) Command {
	return Command{
		OrderID: orderID,
	}
}
