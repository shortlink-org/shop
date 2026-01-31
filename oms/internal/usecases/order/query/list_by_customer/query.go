package list_by_customer

import (
	"github.com/google/uuid"
)

// Query represents a query to list orders by customer ID.
type Query struct {
	CustomerID uuid.UUID
}

// NewQuery creates a new ListOrdersByCustomer query.
func NewQuery(customerID uuid.UUID) Query {
	return Query{
		CustomerID: customerID,
	}
}
