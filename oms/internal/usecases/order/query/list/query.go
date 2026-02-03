package list

import (
	"github.com/google/uuid"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Query is a query to list orders with optional filters.
type Query struct {
	CustomerID   *uuid.UUID
	StatusFilter []order.OrderStatus
}

// NewQuery creates a new ListOrders query.
func NewQuery(customerID *uuid.UUID, statusFilter []order.OrderStatus) Query {
	return Query{
		CustomerID:   customerID,
		StatusFilter: statusFilter,
	}
}
