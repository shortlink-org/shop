package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestTaxNonNegativeSpec(t *testing.T) {
	t.Parallel()

	t.Run("positive tax passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(100), decimal.Zero, decimal.NewFromFloat(8.00))
		assert.NoError(t, err)

		spec := TaxNonNegativeSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("zero tax passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(100), decimal.Zero, decimal.Zero)
		assert.NoError(t, err)

		spec := TaxNonNegativeSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("negative tax fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(100), decimal.Zero, decimal.NewFromFloat(-2.00))
		assert.ErrorIs(t, err, itemv1.ErrItemTaxNegative)
	})
}
