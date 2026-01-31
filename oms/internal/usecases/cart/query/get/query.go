package get

import (
	"github.com/google/uuid"
)

// Query represents a query to get a cart.
type Query struct {
	CustomerID uuid.UUID
}

// NewQuery creates a new GetCart query.
func NewQuery(customerID uuid.UUID) Query {
	return Query{
		CustomerID: customerID,
	}
}
