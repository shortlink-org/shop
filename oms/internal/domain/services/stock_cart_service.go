package services

import (
	"github.com/google/uuid"

	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// StockCartService is a domain service that handles stock-related operations for carts.
// This service contains business logic that involves multiple aggregates (Cart and Stock).
//
// Note: This is a pure domain service - it operates on domain aggregates directly.
// The application layer (UseCase) is responsible for loading/saving aggregates.
type StockCartService struct{}

// NewStockCartService creates a new StockCartService
func NewStockCartService() *StockCartService {
	return &StockCartService{}
}

// StockDepletionResult represents the result of processing a stock depletion for a customer's cart
type StockDepletionResult struct {
	CustomerID uuid.UUID
	GoodID     uuid.UUID
	Removed    bool
	Quantity   int32
	Error      error
}

// ProcessStockDepletion processes stock depletion for a cart.
// This is a pure domain operation - it modifies the cart aggregate in memory.
// The caller (UseCase) is responsible for loading the cart before and saving after.
//
// Pattern: The UseCase does Load -> this domain service method -> Save
func (s *StockCartService) ProcessStockDepletion(
	cartState *cart.State,
	goodID uuid.UUID,
) StockDepletionResult {
	customerID := cartState.GetCustomerId()

	// Find the item in the cart
	var itemQuantity int32
	var found bool

	for _, item := range cartState.GetItems() {
		if item.GetGoodId() == goodID {
			itemQuantity = item.GetQuantity()
			found = true
			break
		}
	}

	if !found {
		// Item was already removed from cart
		return StockDepletionResult{
			CustomerID: customerID,
			GoodID:     goodID,
			Removed:    false,
			Quantity:   0,
		}
	}

	// Create an item to remove (with the quantity to remove)
	itemToRemove, err := itemv1.NewItem(goodID, itemQuantity)
	if err != nil {
		return StockDepletionResult{
			CustomerID: customerID,
			GoodID:     goodID,
			Removed:    false,
			Quantity:   itemQuantity,
			Error:      err,
		}
	}

	// Remove the item from cart (domain method)
	if err := cartState.RemoveItem(itemToRemove); err != nil {
		return StockDepletionResult{
			CustomerID: customerID,
			GoodID:     goodID,
			Removed:    false,
			Quantity:   itemQuantity,
			Error:      err,
		}
	}

	return StockDepletionResult{
		CustomerID: customerID,
		GoodID:     goodID,
		Removed:    true,
		Quantity:   itemQuantity,
	}
}

