package v1

import (
	"fmt"

	"github.com/google/uuid"

	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// ValidateAddItemsWithStock validates if items can be added to a cart using
// pre-fetched stock data. Pure domain logic: no I/O.
// The use case must obtain stockByGoodId via a port (e.g. StockChecker) and pass it here.
func ValidateAddItemsWithStock(
	items itemsv1.Items,
	stockByGoodId map[uuid.UUID]StockAvailabilityInput,
) Result {
	result := Result{
		Valid:    true,
		Errors:   make([]Error, 0),
		Warnings: make([]Warning, 0),
	}

	for _, item := range items {
		goodID := item.GetGoodId()
		stock, hasStock := stockByGoodId[goodID]

		if hasStock && stock.CheckError != nil {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  goodID,
				Message: fmt.Sprintf("Failed to check stock availability: %v", stock.CheckError),
				Code:    "STOCK_CHECK_ERROR",
			})

			continue
		}

		if !hasStock || !stock.Available {
			quantity := uint32(0)
			if hasStock {
				quantity = stock.StockQuantity
			}

			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  goodID,
				Message: fmt.Sprintf("Insufficient stock. Available: %d, Requested: %d", quantity, item.GetQuantity()),
				Code:    "INSUFFICIENT_STOCK",
			})

			continue
		}

		if uint32(item.GetQuantity()) > stock.StockQuantity {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  goodID,
				Message: fmt.Sprintf("Requested quantity (%d) exceeds available stock (%d)", item.GetQuantity(), stock.StockQuantity),
				Code:    "QUANTITY_EXCEEDS_STOCK",
			})
		}

		if item.GetQuantity() <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, Error{
				GoodID:  goodID,
				Message: "Quantity must be greater than zero",
				Code:    "INVALID_QUANTITY",
			})
		}
	}

	return result
}

// ValidateAddItems validates if items can be added to a cart using pre-fetched stock data.
// Pure domain: no I/O. Use case must populate stockByGoodId via a port (e.g. StockChecker).
func (s *Service) ValidateAddItems(
	items itemsv1.Items,
	stockByGoodId map[uuid.UUID]StockAvailabilityInput,
) Result {
	return ValidateAddItemsWithStock(items, stockByGoodId)
}
