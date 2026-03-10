package ports

import (
	"context"

	"github.com/google/uuid"

	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
)

// CartRepository defines the minimal interface for cart persistence.
// Repository is a storage adapter (infrastructure layer), NOT a use-case.
//
// Rules:
//   - Only Load and Save operations (no business operations like AddItem/RemoveItem)
//   - UseCase orchestrates: Load -> domain method(s) -> Save
//   - Domain aggregate contains behavior and invariants
//
//nolint:iface // port interface used by usecases and DI
type CartRepository interface {
	Load(ctx context.Context, customerID uuid.UUID) (*cart.State, error)
	Save(ctx context.Context, state *cart.State) error
}
