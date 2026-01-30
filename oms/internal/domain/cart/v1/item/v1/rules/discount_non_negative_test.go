package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestDiscountNonNegativeSpec(t *testing.T) {
	t.Parallel()

	t.Run("positive discount passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(100), decimal.NewFromFloat(10.00), decimal.Zero)
		assert.NoError(t, err)

		spec := DiscountNonNegativeSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("zero discount passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(100), decimal.Zero, decimal.Zero)
		assert.NoError(t, err)

		spec := DiscountNonNegativeSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("negative discount fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(100), decimal.NewFromFloat(-5.00), decimal.Zero)
		assert.ErrorIs(t, err, itemv1.ErrItemDiscountNegative)
	})
}
