package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestGoodIdNotEmptySpec(t *testing.T) {
	t.Parallel()

	t.Run("valid goodId passes", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(uuid.New(), 1, decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
		assert.NoError(t, err)

		spec := GoodIdNotEmptySpec{}
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})

	t.Run("empty goodId fails in constructor", func(t *testing.T) {
		t.Parallel()

		_, err := itemv1.NewItemWithPricing(uuid.Nil, 1, decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
		assert.ErrorIs(t, err, itemv1.ErrItemGoodIdZero)
	})
}
