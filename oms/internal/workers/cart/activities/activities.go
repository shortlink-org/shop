package activities

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	cartuc "github.com/shortlink-org/shop/oms/internal/usecases/cart"
)

// Activities wraps cart use case methods for Temporal activities.
// Activities are the bridge between Temporal workflows and application use cases.
// Temporal workflows must never access repositories directly - only through activities.
type Activities struct {
	cartUC *cartuc.UC
}

// New creates a new Activities instance.
func New(cartUC *cartuc.UC) *Activities {
	return &Activities{
		cartUC: cartUC,
	}
}

// AddItemRequest represents the request for AddItem activity.
type AddItemRequest struct {
	CustomerID uuid.UUID
	GoodID     uuid.UUID
	Quantity   int32
	Price      decimal.Decimal
	Discount   decimal.Decimal
}

// AddItem adds an item to the cart.
func (a *Activities) AddItem(ctx context.Context, req AddItemRequest) error {
	item, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
	if err != nil {
		return err
	}

	return a.cartUC.Add(ctx, req.CustomerID, item)
}

// RemoveItemRequest represents the request for RemoveItem activity.
type RemoveItemRequest struct {
	CustomerID uuid.UUID
	GoodID     uuid.UUID
	Quantity   int32
}

// RemoveItem removes an item from the cart.
func (a *Activities) RemoveItem(ctx context.Context, req RemoveItemRequest) error {
	item, err := itemv1.NewItem(req.GoodID, req.Quantity)
	if err != nil {
		return err
	}

	return a.cartUC.Remove(ctx, req.CustomerID, item)
}

// ResetCartRequest represents the request for ResetCart activity.
type ResetCartRequest struct {
	CustomerID uuid.UUID
}

// ResetCart resets the cart.
func (a *Activities) ResetCart(ctx context.Context, req ResetCartRequest) error {
	return a.cartUC.Reset(ctx, req.CustomerID)
}
