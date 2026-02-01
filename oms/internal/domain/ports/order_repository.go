package ports

import (
	"context"

	"github.com/google/uuid"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// ListFilter contains optional filters for listing orders.
type ListFilter struct {
	// CustomerID filters by customer (optional)
	CustomerID *uuid.UUID
	// StatusFilter filters by order status (optional)
	StatusFilter []order.OrderStatus
}

// ListResult contains paginated list results.
type ListResult struct {
	Orders     []*order.OrderState
	TotalCount int64
	TotalPages int32
}

// OrderRepository defines the minimal interface for order persistence.
// Repository is a storage adapter (infrastructure layer), NOT a use-case.
//
// Rules:
//   - Only Load and Save operations (no business operations like Cancel/Complete)
//   - UseCase orchestrates: Load -> domain method(s) -> Save
//   - Domain aggregate contains behavior and invariants
type OrderRepository interface {
	// Load retrieves an order by ID.
	// Returns ErrNotFound if the order does not exist.
	Load(ctx context.Context, orderID uuid.UUID) (*order.OrderState, error)

	// Save persists the order state.
	// Implements optimistic concurrency control via version field.
	// Returns ErrVersionConflict if the version has changed since loading.
	Save(ctx context.Context, state *order.OrderState) error

	// ListByCustomer retrieves all orders for a customer.
	// Returns empty slice if no orders exist.
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*order.OrderState, error)

	// List retrieves orders with filtering and pagination.
	// Returns empty slice if no orders match the filter.
	List(ctx context.Context, filter ListFilter, page, pageSize int32) (*ListResult, error)
}
