package ports

import (
	"context"

	"github.com/google/uuid"
)

// CartGoodsIndex provides an index for quickly looking up which customers
// have a specific good in their cart. This is used for stock change notifications.
//
//nolint:iface // port interface used by usecases and DI
type CartGoodsIndex interface {
	AddGoodToCart(ctx context.Context, goodID, customerID uuid.UUID) error
	RemoveGoodFromCart(ctx context.Context, goodID, customerID uuid.UUID) error
	GetCustomersWithGood(ctx context.Context, goodID uuid.UUID) ([]uuid.UUID, error)
}
