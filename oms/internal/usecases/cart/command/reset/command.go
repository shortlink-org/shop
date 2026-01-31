package reset

import (
	"github.com/google/uuid"
)

// Command represents a command to reset (clear) a cart.
type Command struct {
	CustomerID uuid.UUID
}

// NewCommand creates a new Reset command.
func NewCommand(customerID uuid.UUID) Command {
	return Command{
		CustomerID: customerID,
	}
}
