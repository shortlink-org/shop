package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CartValidationService is a domain service that validates cart operations.
// This service contains business rules for cart validation that don't belong to a single entity.
type CartValidationService struct {
	// StockChecker is an interface for checking stock availability
	StockChecker StockChecker
}

// StockChecker defines the interface for checking stock availability
type StockChecker interface {
	// CheckStockAvailability checks if a good has sufficient stock
	CheckStockAvailability(ctx context.Context, goodId uuid.UUID, requestedQuantity int32) (bool, uint32, error)
}

// NewCartValidationService creates a new CartValidationService
func NewCartValidationService(stockChecker StockChecker) *CartValidationService {
	return &CartValidationService{
		StockChecker: stockChecker,
	}
}

// ValidationResult represents the result of a cart validation
type ValidationResult struct {
	Valid   bool
	Errors  []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error
type ValidationError struct {
	GoodID   uuid.UUID
	Message  string
	Code     string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	GoodID   uuid.UUID
	Message  string
	Code     string
}

// ValidateAddItems validates if items can be added to a cart.
// This encapsulates business rules like:
// - Stock availability checks
// - Quantity limits
// - Item restrictions
func (s *CartValidationService) ValidateAddItems(
	ctx context.Context,
	items CartItems,
) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	for _, item := range items {
		// Check stock availability
		available, stockQuantity, err := s.StockChecker.CheckStockAvailability(
			ctx,
			item.GetGoodId(),
			item.GetQuantity(),
		)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Failed to check stock availability: %v", err),
				Code:    "STOCK_CHECK_ERROR",
			})
			continue
		}

		if !available {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", stockQuantity, item.GetQuantity()),
				Code:    "INSUFFICIENT_STOCK",
			})
			continue
		}

		// Check if requested quantity exceeds available stock
		if uint32(item.GetQuantity()) > stockQuantity {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Requested quantity (%d) exceeds available stock (%d)", item.GetQuantity(), stockQuantity),
				Code:    "QUANTITY_EXCEEDS_STOCK",
			})
		}

		// Validate quantity is positive
		if item.GetQuantity() <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				GoodID:  item.GetGoodId(),
				Message: "Quantity must be greater than zero",
				Code:    "INVALID_QUANTITY",
			})
		}
	}

	return result
}

// ValidateRemoveItems validates if items can be removed from a cart.
func (s *CartValidationService) ValidateRemoveItems(
	ctx context.Context,
	cart *CartState,
	items CartItems,
) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	cartItems := cart.GetItems()
	cartItemMap := make(map[uuid.UUID]CartItem)
	for _, item := range cartItems {
		cartItemMap[item.GetGoodId()] = item
	}

	for _, item := range items {
		cartItem, exists := cartItemMap[item.GetGoodId()]
		if !exists {
			result.Warnings = append(result.Warnings, ValidationWarning{
				GoodID:  item.GetGoodId(),
				Message: "Item not found in cart",
				Code:    "ITEM_NOT_IN_CART",
			})
			continue
		}

		// Check if removing more than available
		if item.GetQuantity() > cartItem.GetQuantity() {
			result.Warnings = append(result.Warnings, ValidationWarning{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Removing more items (%d) than in cart (%d)", item.GetQuantity(), cartItem.GetQuantity()),
				Code:    "REMOVE_EXCEEDS_CART",
			})
		}
	}

	return result
}

