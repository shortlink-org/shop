package list

import (
	"github.com/google/uuid"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Query represents a query to list orders with filters and pagination.
type Query struct {
	// CustomerID filters by customer (optional)
	CustomerID *uuid.UUID
	// StatusFilter filters by order status (optional)
	StatusFilter []order.OrderStatus
	// Page number (1-indexed)
	Page int32
	// PageSize is the number of items per page
	PageSize int32
}

// NewQuery creates a new ListOrders query.
func NewQuery(customerID *uuid.UUID, statusFilter []order.OrderStatus, page, pageSize int32) Query {
	// Ensure reasonable defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return Query{
		CustomerID:   customerID,
		StatusFilter: statusFilter,
		Page:         page,
		PageSize:     pageSize,
	}
}
