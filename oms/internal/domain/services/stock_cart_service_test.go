package services

import (
	"testing"

	"github.com/google/uuid"

	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	item "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func createCartWithItem(customerId, goodId uuid.UUID, quantity int32) *cart.State {
	state := cart.New(customerId)
	i, _ := item.NewItem(goodId, quantity)
	_ = state.AddItem(i)

	return state
}

func TestStockCartService_ProcessStockDepletion(t *testing.T) {
	goodId := uuid.New()
	customerId := uuid.New()

	tests := []struct {
		name         string
		setupCart    func() *cart.State
		wantRemoved  bool
		wantQuantity int32
		wantError    bool
	}{
		{
			name: "removes item from cart",
			setupCart: func() *cart.State {
				return createCartWithItem(customerId, goodId, 2)
			},
			wantRemoved:  true,
			wantQuantity: 2,
			wantError:    false,
		},
		{
			name: "handles item not in cart",
			setupCart: func() *cart.State {
				otherGoodId := uuid.New()
				return createCartWithItem(customerId, otherGoodId, 2)
			},
			wantRemoved:  false,
			wantQuantity: 0,
			wantError:    false,
		},
		{
			name: "handles empty cart",
			setupCart: func() *cart.State {
				return cart.New(customerId)
			},
			wantRemoved:  false,
			wantQuantity: 0,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartState := tt.setupCart()
			service := NewStockCartService()

			result := service.ProcessStockDepletion(cartState, goodId)

			if result.Removed != tt.wantRemoved {
				t.Errorf("ProcessStockDepletion() removed = %v, want %v", result.Removed, tt.wantRemoved)
			}

			if result.Quantity != tt.wantQuantity {
				t.Errorf("ProcessStockDepletion() quantity = %d, want %d", result.Quantity, tt.wantQuantity)
			}

			if (result.Error != nil) != tt.wantError {
				t.Errorf("ProcessStockDepletion() error = %v, wantError = %v", result.Error, tt.wantError)
			}

			if result.CustomerID != customerId {
				t.Errorf("ProcessStockDepletion() customerID = %v, want %v", result.CustomerID, customerId)
			}

			if result.GoodID != goodId {
				t.Errorf("ProcessStockDepletion() goodID = %v, want %v", result.GoodID, goodId)
			}
		})
	}
}

func TestStockCartService_ProcessStockDepletion_ModifiesCart(t *testing.T) {
	goodId := uuid.New()
	customerId := uuid.New()

	cartState := createCartWithItem(customerId, goodId, 3)
	service := NewStockCartService()

	// Verify item is in cart before
	itemsBefore := cartState.GetItems()
	if len(itemsBefore) != 1 {
		t.Fatalf("Expected 1 item in cart, got %d", len(itemsBefore))
	}

	// Process stock depletion
	result := service.ProcessStockDepletion(cartState, goodId)

	if !result.Removed {
		t.Error("Expected item to be removed")
	}

	// Verify item is removed from cart after
	itemsAfter := cartState.GetItems()
	if len(itemsAfter) != 0 {
		t.Errorf("Expected 0 items in cart after removal, got %d", len(itemsAfter))
	}
}
