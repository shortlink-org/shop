package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	item "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	items "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// MockStockChecker is a mock implementation of StockChecker
type MockStockChecker struct {
	stocks map[uuid.UUID]uint32
	err    error
}

func NewMockStockChecker() *MockStockChecker {
	return &MockStockChecker{
		stocks: make(map[uuid.UUID]uint32),
	}
}

func (m *MockStockChecker) CheckStockAvailability(ctx context.Context, goodId uuid.UUID, requestedQuantity int32) (bool, uint32, error) {
	if m.err != nil {
		return false, 0, m.err
	}

	stock, ok := m.stocks[goodId]
	if !ok {
		return false, 0, nil
	}

	return uint32(requestedQuantity) <= stock, stock, nil
}

func (m *MockStockChecker) SetStock(goodId uuid.UUID, quantity uint32) {
	m.stocks[goodId] = quantity
}

func mustNewItem(goodId uuid.UUID, quantity int32) item.Item {
	i, err := item.NewItem(goodId, quantity)
	if err != nil {
		panic(err)
	}
	return i
}

func TestService_ValidateAddItems(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupStock func(*MockStockChecker)
		items      items.Items
		wantValid  bool
		wantErrors int
	}{
		{
			name: "valid items with sufficient stock",
			setupStock: func(checker *MockStockChecker) {
				goodId := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				checker.SetStock(goodId, 10)
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("11111111-1111-1111-1111-111111111111"), 5),
			},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "insufficient stock",
			setupStock: func(checker *MockStockChecker) {
				goodId := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				checker.SetStock(goodId, 2)
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("22222222-2222-2222-2222-222222222222"), 5),
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "zero stock",
			setupStock: func(checker *MockStockChecker) {
				goodId := uuid.MustParse("33333333-3333-3333-3333-333333333333")
				checker.SetStock(goodId, 0)
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("33333333-3333-3333-3333-333333333333"), 1),
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name:       "empty items list",
			setupStock: func(checker *MockStockChecker) {},
			items:      items.Items{},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "multiple items - all valid",
			setupStock: func(checker *MockStockChecker) {
				checker.SetStock(uuid.MustParse("44444444-4444-4444-4444-444444444444"), 10)
				checker.SetStock(uuid.MustParse("55555555-5555-5555-5555-555555555555"), 20)
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("44444444-4444-4444-4444-444444444444"), 5),
				mustNewItem(uuid.MustParse("55555555-5555-5555-5555-555555555555"), 10),
			},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "multiple items - one invalid",
			setupStock: func(checker *MockStockChecker) {
				checker.SetStock(uuid.MustParse("66666666-6666-6666-6666-666666666666"), 10)
				checker.SetStock(uuid.MustParse("77777777-7777-7777-7777-777777777777"), 2)
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("66666666-6666-6666-6666-666666666666"), 5),
				mustNewItem(uuid.MustParse("77777777-7777-7777-7777-777777777777"), 10), // Exceeds stock
			},
			wantValid:  false,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewMockStockChecker()
			tt.setupStock(checker)

			service := New(checker)
			result := service.ValidateAddItems(ctx, tt.items)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateAddItems() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("ValidateAddItems() errors = %d, want %d", len(result.Errors), tt.wantErrors)
			}
		})
	}
}

func TestService_ValidateAddItems_StockCheckError(t *testing.T) {
	ctx := context.Background()
	goodId := uuid.New()

	checker := NewMockStockChecker()
	checker.err = errors.New("stock service unavailable")

	service := New(checker)
	result := service.ValidateAddItems(ctx, items.Items{
		mustNewItem(goodId, 1),
	})

	if result.Valid {
		t.Error("ValidateAddItems() should be invalid when stock check fails")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAddItems() errors = %d, want 1", len(result.Errors))
	}

	if result.Errors[0].Code != "STOCK_CHECK_ERROR" {
		t.Errorf("ValidateAddItems() error code = %s, want STOCK_CHECK_ERROR", result.Errors[0].Code)
	}
}
