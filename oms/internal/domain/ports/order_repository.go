package ports

import (
	"context"

	"github.com/google/uuid"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// OrderRepository defines the minimal interface for order persistence.
// Repository is a storage adapter (infrastructure layer), NOT a use-case.
//
// Rules:
//   - Only Load and Save operations (no business operations like Cancel/Complete)
//   - UseCase orchestrates: Load -> domain method(s) -> Save
//   - Domain aggregate contains behavior and invariants
type OrderRepository interface {
	Load(ctx context.Context, orderID uuid.UUID) (*order.OrderState, error)
	Save(ctx context.Context, state *order.OrderState) error
	List(ctx context.Context, filter ListFilter) ([]*order.OrderState, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*order.OrderState, error)
}
