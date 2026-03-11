package get

import (
	"github.com/google/uuid"
)

// Query represents a query to get an order by ID.
type Query struct {
	OrderID    uuid.UUID
	CustomerID *uuid.UUID
}

// NewQuery creates a new GetOrder query.
func NewQuery(orderID uuid.UUID) Query {
	return Query{
		OrderID: orderID,
	}
}

// NewCustomerScopedQuery creates a query constrained to a specific customer.
func NewCustomerScopedQuery(orderID, customerID uuid.UUID) Query {
	return Query{
		OrderID:    orderID,
		CustomerID: &customerID,
	}
}
