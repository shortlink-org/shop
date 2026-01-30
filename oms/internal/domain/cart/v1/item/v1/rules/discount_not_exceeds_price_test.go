package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestDiscountNotExceedsPriceSpec(t *testing.T) {
	t.Parallel()

	t.Run("discount less than price passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromFloat(100.00), decimal.NewFromFloat(20.00), decimal.Zero)
		assert.NoError(t, err)

		spec := DiscountNotExceedsPriceSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("discount equals price passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromFloat(100.00), decimal.NewFromFloat(100.00), decimal.Zero)
		assert.NoError(t, err)

		spec := DiscountNotExceedsPriceSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("discount exceeds price fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromFloat(100.00), decimal.NewFromFloat(150.00), decimal.Zero)
		assert.ErrorIs(t, err, itemv1.ErrItemDiscountExceedsPrice)
	})
}
