package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	item "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// MockCartRepository is a mock implementation of CartRepository
type MockCartRepository struct {
	carts       map[uuid.UUID]*cart.State
	removeError error
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		carts: make(map[uuid.UUID]*cart.State),
	}
}

func (m *MockCartRepository) GetCart(ctx context.Context, customerId uuid.UUID) (*cart.State, error) {
	if c, ok := m.carts[customerId]; ok {
		return c, nil
	}
	return nil, errors.New("cart not found")
}

func (m *MockCartRepository) RemoveItemFromCart(ctx context.Context, customerId uuid.UUID, goodId uuid.UUID, quantity int32) error {
	if m.removeError != nil {
		return m.removeError
	}
	return nil
}

func (m *MockCartRepository) AddCart(customerId uuid.UUID, goodId uuid.UUID, quantity int32) {
	state := cart.New(customerId)
	i, _ := item.NewItem(goodId, quantity)
	state.AddItem(i)
	m.carts[customerId] = state
}

func TestStockCartService_HandleStockDepletion(t *testing.T) {
	ctx := context.Background()
	goodId := uuid.New()
	customerId1 := uuid.New()
	customerId2 := uuid.New()

	tests := []struct {
		name                string
		setupRepo           func(*MockCartRepository)
		affectedCustomerIds []uuid.UUID
		wantResults         int
		wantRemovedCount    int
	}{
		{
			name: "removes item from affected carts",
			setupRepo: func(repo *MockCartRepository) {
				repo.AddCart(customerId1, goodId, 2)
				repo.AddCart(customerId2, goodId, 1)
			},
			affectedCustomerIds: []uuid.UUID{customerId1, customerId2},
			wantResults:         2,
			wantRemovedCount:    2,
		},
		{
			name: "handles cart not found",
			setupRepo: func(repo *MockCartRepository) {
				// Don't add any carts
			},
			affectedCustomerIds: []uuid.UUID{customerId1},
			wantResults:         1,
			wantRemovedCount:    0,
		},
		{
			name: "handles item not in cart",
			setupRepo: func(repo *MockCartRepository) {
				otherGoodId := uuid.New()
				repo.AddCart(customerId1, otherGoodId, 2) // Different good
			},
			affectedCustomerIds: []uuid.UUID{customerId1},
			wantResults:         1,
			wantRemovedCount:    0,
		},
		{
			name: "handles empty affected list",
			setupRepo: func(repo *MockCartRepository) {
				repo.AddCart(customerId1, goodId, 2)
			},
			affectedCustomerIds: []uuid.UUID{},
			wantResults:         0,
			wantRemovedCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockCartRepository()
			tt.setupRepo(repo)

			service := NewStockCartService(repo)
			results, err := service.HandleStockDepletion(ctx, goodId, tt.affectedCustomerIds)

			if err != nil {
				t.Errorf("HandleStockDepletion() error = %v", err)
				return
			}

			if len(results) != tt.wantResults {
				t.Errorf("HandleStockDepletion() got %d results, want %d", len(results), tt.wantResults)
			}

			removedCount := 0
			for _, r := range results {
				if r.Removed {
					removedCount++
				}
			}

			if removedCount != tt.wantRemovedCount {
				t.Errorf("HandleStockDepletion() removed %d items, want %d", removedCount, tt.wantRemovedCount)
			}
		})
	}
}

func TestStockCartService_HandleStockDepletion_RemoveError(t *testing.T) {
	ctx := context.Background()
	goodId := uuid.New()
	customerId := uuid.New()

	repo := NewMockCartRepository()
	repo.AddCart(customerId, goodId, 2)
	repo.removeError = errors.New("remove failed")

	service := NewStockCartService(repo)
	results, err := service.HandleStockDepletion(ctx, goodId, []uuid.UUID{customerId})

	if err != nil {
		t.Errorf("HandleStockDepletion() unexpected error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("HandleStockDepletion() got %d results, want 1", len(results))
	}

	if results[0].Removed {
		t.Error("HandleStockDepletion() should not mark as removed when remove fails")
	}

	if results[0].Error == nil {
		t.Error("HandleStockDepletion() should have error when remove fails")
	}
}
