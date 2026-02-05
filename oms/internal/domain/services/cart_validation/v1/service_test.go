package v1

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	item "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	items "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

func mustNewItem(goodId uuid.UUID, quantity int32) item.Item {
	i, err := item.NewItem(goodId, quantity)
	if err != nil {
		panic(err)
	}

	return i
}

func TestValidateAddItemsWithStock(t *testing.T) {
	tests := []struct {
		name        string
		stockByGood map[uuid.UUID]StockAvailabilityInput
		items       items.Items
		wantValid   bool
		wantErrors  int
		wantErrCode string
	}{
		{
			name: "valid items with sufficient stock",
			stockByGood: map[uuid.UUID]StockAvailabilityInput{
				uuid.MustParse("11111111-1111-1111-1111-111111111111"): {
					GoodID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Available:     true,
					StockQuantity: 10,
				},
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("11111111-1111-1111-1111-111111111111"), 5),
			},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "insufficient stock",
			stockByGood: map[uuid.UUID]StockAvailabilityInput{
				uuid.MustParse("22222222-2222-2222-2222-222222222222"): {
					GoodID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					Available:     true,
					StockQuantity: 2,
				},
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("22222222-2222-2222-2222-222222222222"), 5),
			},
			wantValid:   false,
			wantErrors:  1,
			wantErrCode: "QUANTITY_EXCEEDS_STOCK",
		},
		{
			name: "zero stock",
			stockByGood: map[uuid.UUID]StockAvailabilityInput{
				uuid.MustParse("33333333-3333-3333-3333-333333333333"): {
					GoodID:        uuid.MustParse("33333333-3333-3333-3333-333333333333"),
					Available:     true,
					StockQuantity: 0,
				},
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("33333333-3333-3333-3333-333333333333"), 1),
			},
			wantValid:   false,
			wantErrors:  1,
			wantErrCode: "QUANTITY_EXCEEDS_STOCK",
		},
		{
			name:        "empty items list",
			stockByGood: map[uuid.UUID]StockAvailabilityInput{},
			items:       items.Items{},
			wantValid:   true,
			wantErrors:  0,
		},
		{
			name: "multiple items - all valid",
			stockByGood: map[uuid.UUID]StockAvailabilityInput{
				uuid.MustParse("44444444-4444-4444-4444-444444444444"): {
					GoodID:        uuid.MustParse("44444444-4444-4444-4444-444444444444"),
					Available:     true,
					StockQuantity: 10,
				},
				uuid.MustParse("55555555-5555-5555-5555-555555555555"): {
					GoodID:        uuid.MustParse("55555555-5555-5555-5555-555555555555"),
					Available:     true,
					StockQuantity: 20,
				},
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
			stockByGood: map[uuid.UUID]StockAvailabilityInput{
				uuid.MustParse("66666666-6666-6666-6666-666666666666"): {
					GoodID:        uuid.MustParse("66666666-6666-6666-6666-666666666666"),
					Available:     true,
					StockQuantity: 10,
				},
				uuid.MustParse("77777777-7777-7777-7777-777777777777"): {
					GoodID:        uuid.MustParse("77777777-7777-7777-7777-777777777777"),
					Available:     true,
					StockQuantity: 2,
				},
			},
			items: items.Items{
				mustNewItem(uuid.MustParse("66666666-6666-6666-6666-666666666666"), 5),
				mustNewItem(uuid.MustParse("77777777-7777-7777-7777-777777777777"), 10),
			},
			wantValid:   false,
			wantErrors:  1,
			wantErrCode: "QUANTITY_EXCEEDS_STOCK",
		},
		{
			name:        "good not in stock map - insufficient",
			stockByGood: map[uuid.UUID]StockAvailabilityInput{},
			items: items.Items{
				mustNewItem(uuid.MustParse("88888888-8888-8888-8888-888888888888"), 1),
			},
			wantValid:   false,
			wantErrors:  1,
			wantErrCode: "INSUFFICIENT_STOCK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAddItemsWithStock(tt.items, tt.stockByGood)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateAddItemsWithStock() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("ValidateAddItemsWithStock() errors = %d, want %d", len(result.Errors), tt.wantErrors)
			}

			if tt.wantErrCode != "" && len(result.Errors) > 0 && result.Errors[0].Code != tt.wantErrCode {
				t.Errorf("ValidateAddItemsWithStock() error code = %s, want %s", result.Errors[0].Code, tt.wantErrCode)
			}
		})
	}
}

func TestValidateAddItemsWithStock_StockCheckError(t *testing.T) {
	goodId := uuid.New()
	stockByGood := map[uuid.UUID]StockAvailabilityInput{
		goodId: {
			GoodID:     goodId,
			Available:  false,
			CheckError: errors.New("stock service unavailable"),
		},
	}

	result := ValidateAddItemsWithStock(items.Items{mustNewItem(goodId, 1)}, stockByGood)

	if result.Valid {
		t.Error("ValidateAddItemsWithStock() should be invalid when stock check fails")
	}

	if len(result.Errors) != 1 {
		t.Errorf("ValidateAddItemsWithStock() errors = %d, want 1", len(result.Errors))
	}

	if result.Errors[0].Code != "STOCK_CHECK_ERROR" {
		t.Errorf("ValidateAddItemsWithStock() error code = %s, want STOCK_CHECK_ERROR", result.Errors[0].Code)
	}
}

func TestService_ValidateAddItems(t *testing.T) {
	service := New()
	goodId := uuid.New()
	stockByGood := map[uuid.UUID]StockAvailabilityInput{
		goodId: {GoodID: goodId, Available: true, StockQuantity: 5},
	}

	result := service.ValidateAddItems(items.Items{mustNewItem(goodId, 2)}, stockByGood)
	if !result.Valid {
		t.Errorf("Service.ValidateAddItems() valid = false, want true")
	}
}
