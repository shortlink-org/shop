package activities

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.temporal.io/sdk/temporal"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	add_item "github.com/shortlink-org/shop/oms/internal/usecases/cart/command/add_item"
	remove_item "github.com/shortlink-org/shop/oms/internal/usecases/cart/command/remove_item"
	reset "github.com/shortlink-org/shop/oms/internal/usecases/cart/command/reset"
)

// Activities wraps cart command handlers for Temporal activities.
// Activities are the bridge between Temporal workflows and application use cases.
// Temporal workflows must never access repositories directly - only through activities.
type Activities struct {
	addItemHandler    *add_item.Handler
	removeItemHandler *remove_item.Handler
	resetHandler      *reset.Handler
}

const cartValidationErrorType = "CartValidationError"

// New creates a new Activities instance.
func New(
	addItemHandler *add_item.Handler,
	removeItemHandler *remove_item.Handler,
	resetHandler *reset.Handler,
) *Activities {
	return &Activities{
		addItemHandler:    addItemHandler,
		removeItemHandler: removeItemHandler,
		resetHandler:      resetHandler,
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
		return mapCartActivityError(err)
	}

	cmd := add_item.NewCommand(req.CustomerID, item)

	return mapCartActivityError(a.addItemHandler.Handle(ctx, cmd))
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
		return mapCartActivityError(err)
	}

	cmd := remove_item.NewCommand(req.CustomerID, item)

	return mapCartActivityError(a.removeItemHandler.Handle(ctx, cmd))
}

// ResetCartRequest represents the request for ResetCart activity.
type ResetCartRequest struct {
	CustomerID uuid.UUID
}

// ResetCart resets the cart.
func (a *Activities) ResetCart(ctx context.Context, req ResetCartRequest) error {
	cmd := reset.NewCommand(req.CustomerID)
	return a.resetHandler.Handle(ctx, cmd)
}

func mapCartActivityError(err error) error {
	if err == nil {
		return nil
	}

	if isCartValidationError(err) {
		return temporal.NewNonRetryableApplicationError(err.Error(), cartValidationErrorType, err)
	}

	return err
}

func isCartValidationError(err error) bool {
	return errors.Is(err, itemv1.ErrItemGoodIdZero) ||
		errors.Is(err, itemv1.ErrItemQuantityZero) ||
		errors.Is(err, itemv1.ErrItemPriceNegative) ||
		errors.Is(err, itemv1.ErrItemDiscountNegative) ||
		errors.Is(err, itemv1.ErrItemTaxNegative) ||
		errors.Is(err, itemv1.ErrItemDiscountExceedsPrice)
}
