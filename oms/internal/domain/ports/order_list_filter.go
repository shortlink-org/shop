package ports

import (
	"github.com/google/uuid"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// ListFilter contains optional filters for listing orders (query/port, not domain).
type ListFilter struct {
	CustomerID   *uuid.UUID
	StatusFilter []order.OrderStatus
}
