package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestPriceNonNegativeSpec(t *testing.T) {
	t.Parallel()

	t.Run("positive price passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromFloat(10.99), decimal.Zero, decimal.Zero)
		assert.NoError(t, err)

		spec := PriceNonNegativeSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("zero price passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.Zero, decimal.Zero, decimal.Zero)
		assert.NoError(t, err)

		spec := PriceNonNegativeSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("negative price fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromFloat(-5.00), decimal.Zero, decimal.Zero)
		assert.ErrorIs(t, err, itemv1.ErrItemPriceNegative)
	})
}
