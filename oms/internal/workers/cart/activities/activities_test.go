package activities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// Fixed UUIDs for consistent testing
var (
	testCustomerID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174100")
	testGoodID     = uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
)

func TestAddItemRequest_ItemCreation(t *testing.T) {
	t.Run("ValidRequest", func(t *testing.T) {
		req := AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   2,
			Price:      decimal.NewFromFloat(19.99),
			Discount:   decimal.NewFromFloat(2.00),
		}

		// Simulate what the activity does internally
		item, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
		require.NoError(t, err)
		require.Equal(t, req.GoodID, item.GetGoodId())
		require.Equal(t, req.Quantity, item.GetQuantity())
		require.True(t, req.Price.Equal(item.GetPrice()))
		require.True(t, req.Discount.Equal(item.GetDiscount()))
	})

	t.Run("InvalidGoodID", func(t *testing.T) {
		req := AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     uuid.Nil, // Invalid
			Quantity:   1,
			Price:      decimal.NewFromFloat(10.00),
			Discount:   decimal.Zero,
		}

		_, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemGoodIdZero)
	})

	t.Run("InvalidQuantity", func(t *testing.T) {
		req := AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   0, // Invalid
			Price:      decimal.NewFromFloat(10.00),
			Discount:   decimal.Zero,
		}

		_, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemQuantityZero)
	})

	t.Run("NegativePrice", func(t *testing.T) {
		req := AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   1,
			Price:      decimal.NewFromFloat(-10.00), // Invalid
			Discount:   decimal.Zero,
		}

		_, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemPriceNegative)
	})

	t.Run("NegativeDiscount", func(t *testing.T) {
		req := AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   1,
			Price:      decimal.NewFromFloat(10.00),
			Discount:   decimal.NewFromFloat(-2.00), // Invalid
		}

		_, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemDiscountNegative)
	})

	t.Run("DiscountExceedsPrice", func(t *testing.T) {
		req := AddItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   1,
			Price:      decimal.NewFromFloat(10.00),
			Discount:   decimal.NewFromFloat(15.00), // Invalid - exceeds price
		}

		_, err := itemv1.NewItemWithPricing(req.GoodID, req.Quantity, req.Price, req.Discount, decimal.Zero)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemDiscountExceedsPrice)
	})
}

func TestRemoveItemRequest_ItemCreation(t *testing.T) {
	t.Run("ValidRequest", func(t *testing.T) {
		req := RemoveItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   1,
		}

		// Simulate what the activity does internally
		item, err := itemv1.NewItem(req.GoodID, req.Quantity)
		require.NoError(t, err)
		require.Equal(t, req.GoodID, item.GetGoodId())
		require.Equal(t, req.Quantity, item.GetQuantity())
	})

	t.Run("InvalidGoodID", func(t *testing.T) {
		req := RemoveItemRequest{
			CustomerID: testCustomerID,
			GoodID:     uuid.Nil, // Invalid
			Quantity:   1,
		}

		_, err := itemv1.NewItem(req.GoodID, req.Quantity)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemGoodIdZero)
	})

	t.Run("InvalidQuantity", func(t *testing.T) {
		req := RemoveItemRequest{
			CustomerID: testCustomerID,
			GoodID:     testGoodID,
			Quantity:   -1, // Invalid
		}

		_, err := itemv1.NewItem(req.GoodID, req.Quantity)
		require.Error(t, err)
		require.ErrorIs(t, err, itemv1.ErrItemQuantityZero)
	})
}

func TestResetCartRequest(t *testing.T) {
	t.Run("ValidRequest", func(t *testing.T) {
		req := ResetCartRequest{
			CustomerID: testCustomerID,
		}

		// ResetCart doesn't create items, just verify the request structure
		require.NotEqual(t, uuid.Nil, req.CustomerID)
	})

	t.Run("CustomerIDPresent", func(t *testing.T) {
		customerID := uuid.New()
		req := ResetCartRequest{
			CustomerID: customerID,
		}

		require.Equal(t, customerID, req.CustomerID)
	})
}

func TestActivities_New(t *testing.T) {
	// Test that New returns a valid Activities instance (even with nil handlers)
	// This verifies the constructor works correctly
	activities := New(nil, nil, nil)
	require.NotNil(t, activities)
}

func TestAddItemRequest_Fields(t *testing.T) {
	req := AddItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   5,
		Price:      decimal.NewFromFloat(99.99),
		Discount:   decimal.NewFromFloat(10.00),
	}

	require.Equal(t, testCustomerID, req.CustomerID)
	require.Equal(t, testGoodID, req.GoodID)
	require.Equal(t, int32(5), req.Quantity)
	require.True(t, decimal.NewFromFloat(99.99).Equal(req.Price))
	require.True(t, decimal.NewFromFloat(10.00).Equal(req.Discount))
}

func TestRemoveItemRequest_Fields(t *testing.T) {
	req := RemoveItemRequest{
		CustomerID: testCustomerID,
		GoodID:     testGoodID,
		Quantity:   3,
	}

	require.Equal(t, testCustomerID, req.CustomerID)
	require.Equal(t, testGoodID, req.GoodID)
	require.Equal(t, int32(3), req.Quantity)
}

func TestResetCartRequest_Fields(t *testing.T) {
	req := ResetCartRequest{
		CustomerID: testCustomerID,
	}

	require.Equal(t, testCustomerID, req.CustomerID)
}
