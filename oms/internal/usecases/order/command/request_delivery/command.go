package request_delivery

import (
	"time"

	"github.com/google/uuid"
)

// Command records that OMS successfully requested delivery and received a package ID.
type Command struct {
	OrderID     uuid.UUID
	PackageID   uuid.UUID
	RequestedAt time.Time
}

// NewCommand creates a new RequestDelivery command.
func NewCommand(orderID, packageID uuid.UUID, requestedAt time.Time) Command {
	return Command{
		OrderID:     orderID,
		PackageID:   packageID,
		RequestedAt: requestedAt,
	}
}
