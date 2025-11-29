package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	cartDomain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
)

// OrderCreationService is a domain service that handles the creation of orders from carts.
// This service encapsulates the business logic for converting a cart to an order,
// which involves multiple aggregates (Cart and Order).
type OrderCreationService struct {
	// CartRepository is an interface for cart operations
	CartRepository CartRepository
	// StockChecker is an interface for checking stock availability
	StockChecker StockChecker
}

// CartRepository defines the interface for cart repository operations
type CartRepository interface {
	GetCart(ctx context.Context, customerId uuid.UUID) (*cartDomain.CartState, error)
}

// StockChecker defines the interface for checking stock availability
type StockChecker interface {
	CheckStockAvailability(ctx context.Context, goodId uuid.UUID, requestedQuantity int32) (bool, uint32, error)
}

// NewOrderCreationService creates a new OrderCreationService
func NewOrderCreationService(
	cartRepo CartRepository,
	stockChecker StockChecker,
) *OrderCreationService {
	return &OrderCreationService{
		CartRepository: cartRepo,
		StockChecker:    stockChecker,
	}
}

// CreateOrderFromCartResult represents the result of creating an order from a cart
type CreateOrderFromCartResult struct {
	OrderID    uuid.UUID
	OrderItems Items
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

// CreateOrderFromCart creates an order from a customer's cart.
// This method encapsulates the business logic:
// - Validates cart is not empty
// - Checks stock availability for all items
// - Converts cart items to order items
// - Applies business rules for order creation
func (s *OrderCreationService) CreateOrderFromCart(
	ctx context.Context,
	customerId uuid.UUID,
) (*CreateOrderFromCartResult, error) {
	// Get the cart
	cart, err := s.CartRepository.GetCart(ctx, customerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Validate cart is not empty
	cartItems := cart.GetItems()
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cannot create order from empty cart")
	}

	result := &CreateOrderFromCartResult{
		OrderID:    uuid.New(),
		OrderItems: make(Items, 0),
		Errors:     make([]OrderCreationError, 0),
		Warnings:   make([]OrderCreationWarning, 0),
	}

	// Process each cart item
	for _, cartItem := range cartItems {
		// Check stock availability
		available, stockQuantity, err := s.StockChecker.CheckStockAvailability(
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

		if !available {
			result.Errors = append(result.Errors, OrderCreationError{
				GoodID:  cartItem.GetGoodId(),
				Message: fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", stockQuantity, cartItem.GetQuantity()),
				Code:    "INSUFFICIENT_STOCK",
			})
			continue
		}

		// Track the quantity to use (may be adjusted if stock is insufficient)
		quantityToOrder := cartItem.GetQuantity()

		// If stock is less than requested, add warning and adjust quantity
		if uint32(cartItem.GetQuantity()) > stockQuantity {
			result.Warnings = append(result.Warnings, OrderCreationWarning{
				GoodID:  cartItem.GetGoodId(),
				Message: fmt.Sprintf("Stock reduced. Available: %d, Requested: %d. Adjusting quantity", stockQuantity, cartItem.GetQuantity()),
				Code:    "STOCK_REDUCED",
			})
			// Adjust quantity to available stock
			quantityToOrder = int32(stockQuantity)
		}

		// Convert cart item to order item
		// Note: Price should be set from cart or fetched from pricing service
		orderItem := NewItem(
			cartItem.GetGoodId(),
			quantityToOrder,
			cartItem.GetPrice(),
		)

		result.OrderItems = append(result.OrderItems, *orderItem)
	}

	// If there are critical errors, return error
	if len(result.Errors) > 0 {
		return result, fmt.Errorf("order creation failed with %d errors", len(result.Errors))
	}

	return result, nil
}

