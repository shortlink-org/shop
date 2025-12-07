package v1

import (
	"context"
	"fmt"

	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// ValidateAddItems validates if items can be added to a cart.
// This encapsulates business rules like:
// - Stock availability checks
// - Quantity limits
// - Item restrictions
func (s *Service) ValidateAddItems(
	ctx context.Context,
	items itemsv1.Items,
) Result {
	result := Result{
		Valid:    true,
		Errors:   make([]Error, 0),
		Warnings: make([]Warning, 0),
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
			result.Errors = append(result.Errors, Error{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Failed to check stock availability: %v", err),
				Code:    "STOCK_CHECK_ERROR",
			})
			continue
		}

		if !available {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", stockQuantity, item.GetQuantity()),
				Code:    "INSUFFICIENT_STOCK",
			})
			continue
		}

		// Check if requested quantity exceeds available stock
		if uint32(item.GetQuantity()) > stockQuantity {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  item.GetGoodId(),
				Message: fmt.Sprintf("Requested quantity (%d) exceeds available stock (%d)", item.GetQuantity(), stockQuantity),
				Code:    "QUANTITY_EXCEEDS_STOCK",
			})
		}

		// Validate quantity is positive
		if item.GetQuantity() <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  item.GetGoodId(),
				Message: "Quantity must be greater than zero",
				Code:    "INVALID_QUANTITY",
			})
		}
	}

	return result
}

