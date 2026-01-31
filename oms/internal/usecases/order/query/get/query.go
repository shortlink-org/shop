package get

import (
	"github.com/google/uuid"
)

// Query represents a query to get an order by ID.
type Query struct {
	OrderID uuid.UUID
}

// NewQuery creates a new GetOrder query.
func NewQuery(orderID uuid.UUID) Query {
	return Query{
		OrderID: orderID,
	}
}
