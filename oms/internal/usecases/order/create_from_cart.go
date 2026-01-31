package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	cartDomain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	orderDomain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// StockChecker defines the interface for checking stock availability.
// This interface should be implemented by the stock service adapter.
type StockChecker interface {
	CheckStockAvailability(ctx context.Context, goodId uuid.UUID, requestedQuantity int32) (bool, uint32, error)
}

// CreateOrderFromCartResult represents the result of creating an order from a cart
type CreateOrderFromCartResult struct {
	OrderID    uuid.UUID
	OrderItems orderDomain.Items
	Errors     []OrderCreationError
	Warnings   []OrderCreationWarning
}

// OrderCreationError represents an error during order creation
type OrderCreationError struct {
	GoodID  uuid.UUID
	Message string
	Code    string
}

// OrderCreationWarning represents a warning during order creation
type OrderCreationWarning struct {
	GoodID  uuid.UUID
	Message string
	Code    string
}

// CreateFromCart creates an order from a customer's cart.
// This is an application layer service that orchestrates:
// - Stock availability checking
// - Domain logic for converting cart to order
// - Business rules application
//
// Note: The cart should be loaded by the caller (using CartRepository.Load)
// and passed directly. This follows the pattern where UseCase receives
// already-loaded aggregates when orchestrating across bounded contexts.
func (uc *UC) CreateFromCart(
	ctx context.Context,
	cart *cartDomain.State,
	stockChecker StockChecker,
) (*CreateOrderFromCartResult, error) {
	// Validate cart is not empty
	cartItems := cart.GetItems()
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cannot create order from empty cart")
	}

	result := &CreateOrderFromCartResult{
		OrderID:    uuid.New(),
		OrderItems: make(orderDomain.Items, 0),
		Errors:     make([]OrderCreationError, 0),
		Warnings:   make([]OrderCreationWarning, 0),
	}

	// Process each cart item using the formalized stock checking contract
	for _, cartItem := range cartItems {
		// Check stock availability
		available, stockQuantity, err := stockChecker.CheckStockAvailability(
			ctx,
			cartItem.GetGoodId(),
			cartItem.GetQuantity(),
		)
		if err != nil {
			result.Errors = append(result.Errors, OrderCreationError{
				GoodID:  cartItem.GetGoodId(),
				Message: fmt.Sprintf("Failed to check stock: %v", err),
				Code:    "STOCK_CHECK_ERROR",
			})
			continue
		}

		// Use formalized contract to interpret stock availability
		decision := orderDomain.InterpretStockResult(
			available,
			stockQuantity,
			cartItem.GetQuantity(),
			cartItem.GetGoodId(),
		)

		// Handle decision according to business rules
		if decision.IsError() {
			var code string
			switch decision.Type {
			case orderDomain.StockDecisionTypeSKUUnavailable:
				code = "ITEM_NOT_AVAILABLE"
			case orderDomain.StockDecisionTypeOutOfStock:
				code = "NO_STOCK"
			default:
				code = "STOCK_ERROR"
			}

			result.Errors = append(result.Errors, OrderCreationError{
				GoodID:  cartItem.GetGoodId(),
				Message: decision.Reason,
				Code:    code,
			})
			continue
		}

		// Add warning if partial availability (partial stock)
		if decision.IsWarning() {
			result.Warnings = append(result.Warnings, OrderCreationWarning{
				GoodID:  cartItem.GetGoodId(),
				Message: decision.Reason,
				Code:    "STOCK_REDUCED",
			})
		}

		// Convert cart item to order item with determined quantity
		// Note: Price should be set from cart or fetched from pricing service
		orderItem := orderDomain.NewItem(
			cartItem.GetGoodId(),
			decision.QuantityToUse,
			cartItem.GetPrice(),
		)

		result.OrderItems = append(result.OrderItems, orderItem)
	}

	// If there are critical errors, return error
	if len(result.Errors) > 0 {
		return result, fmt.Errorf("order creation failed with %d errors", len(result.Errors))
	}

	return result, nil
}
