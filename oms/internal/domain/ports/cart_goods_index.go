package ports

import (
	"context"

	"github.com/google/uuid"
)

// CartGoodsIndex provides an index for quickly looking up which customers
// have a specific good in their cart. This is used for stock change notifications.
type CartGoodsIndex interface {
	// AddGoodToCart adds a good to a customer's cart in the index.
	AddGoodToCart(ctx context.Context, goodID, customerID uuid.UUID) error

	// RemoveGoodFromCart removes a good from a customer's cart in the index.
	RemoveGoodFromCart(ctx context.Context, goodID, customerID uuid.UUID) error

	// GetCustomersWithGood returns all customer IDs that have the specified good in their cart.
	GetCustomersWithGood(ctx context.Context, goodID uuid.UUID) ([]uuid.UUID, error)

	// ClearCart removes all goods for a customer from the index.
	ClearCart(ctx context.Context, customerID uuid.UUID) error
}
