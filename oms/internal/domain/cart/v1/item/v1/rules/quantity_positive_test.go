package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestQuantityPositiveSpec(t *testing.T) {
	t.Parallel()

	t.Run("positive quantity passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 5, decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
		assert.NoError(t, err)

		spec := QuantityPositiveSpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("zero quantity fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.New(), 0, decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
		assert.ErrorIs(t, err, itemv1.ErrItemQuantityZero)
	})

	t.Run("negative quantity fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.New(), -1, decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
		assert.ErrorIs(t, err, itemv1.ErrItemQuantityZero)
	})
}
