package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// StockCartService is a domain service that handles stock-related operations for carts.
// This service contains business logic that involves multiple aggregates (Cart and Stock).
type StockCartService struct {
	// CartRepository is an interface for cart operations
	// In a real implementation, this would be injected via dependency injection
	CartRepository CartRepository
}

// CartRepository defines the interface for cart repository operations needed by the domain service
type CartRepository interface {
	// GetCart retrieves a cart by customer ID
	GetCart(ctx context.Context, customerId uuid.UUID) (*CartState, error)
	// RemoveItemFromCart removes an item from a cart
	RemoveItemFromCart(ctx context.Context, customerId uuid.UUID, goodId uuid.UUID, quantity int32) error
}

// NewStockCartService creates a new StockCartService
func NewStockCartService(repo CartRepository) *StockCartService {
	return &StockCartService{
		CartRepository: repo,
	}
}

// HandleStockDepletion handles the business logic when stock for a good is depleted.
// This method encapsulates the domain rule: "When stock reaches zero, remove the item from all affected carts"
func (s *StockCartService) HandleStockDepletion(
	ctx context.Context,
	goodId uuid.UUID,
	affectedCustomerIds []uuid.UUID,
) ([]StockDepletionResult, error) {
	var results []StockDepletionResult

	for _, customerId := range affectedCustomerIds {
		result := s.processCustomerCart(ctx, customerId, goodId)
		results = append(results, result)
	}

	return results, nil
}

// StockDepletionResult represents the result of processing a stock depletion for a customer's cart
type StockDepletionResult struct {
	CustomerID uuid.UUID
	GoodID     uuid.UUID
	Removed    bool
	Quantity   int32
	Error      error
}

// processCustomerCart processes stock depletion for a single customer's cart
func (s *StockCartService) processCustomerCart(
	ctx context.Context,
	customerId uuid.UUID,
	goodId uuid.UUID,
) StockDepletionResult {
	// Get the cart
	cart, err := s.CartRepository.GetCart(ctx, customerId)
	if err != nil {
		return StockDepletionResult{
			CustomerID: customerId,
			GoodID:     goodId,
			Removed:    false,
			Error:      fmt.Errorf("failed to get cart: %w", err),
		}
	}

	// Find the item in the cart
	var itemQuantity int32
	found := false
	for _, item := range cart.GetItems() {
		if item.GetGoodId() == goodId {
			itemQuantity = item.GetQuantity()
			found = true
			break
		}
	}

	if !found {
		// Item was already removed from cart
		return StockDepletionResult{
			CustomerID: customerId,
			GoodID:     goodId,
			Removed:    false,
			Quantity:   0,
		}
	}

	// Remove the item from cart
	err = s.CartRepository.RemoveItemFromCart(ctx, customerId, goodId, itemQuantity)
	if err != nil {
		return StockDepletionResult{
			CustomerID: customerId,
			GoodID:     goodId,
			Removed:    false,
			Quantity:   itemQuantity,
			Error:      fmt.Errorf("failed to remove item from cart: %w", err),
		}
	}

	return StockDepletionResult{
		CustomerID: customerId,
		GoodID:     goodId,
		Removed:    true,
		Quantity:   itemQuantity,
	}
}

