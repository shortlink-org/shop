package rules

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

func TestNewItemSpecification(t *testing.T) {
	t.Parallel()

	t.Run("valid item passes all specs", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItem(uuid.New(), 1)
		assert.NoError(t, err)

		spec := NewItemSpecification()
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})
}

func TestNewItemWithPricingSpecification(t *testing.T) {
	t.Parallel()

	t.Run("valid item with pricing passes all specs", func(t *testing.T) {
		t.Parallel()

		item, err := itemv1.NewItemWithPricing(
			uuid.New(),
			2,
			decimal.NewFromFloat(99.99),
			decimal.NewFromFloat(10.00),
			decimal.NewFromFloat(8.00),
		)
		assert.NoError(t, err)

		spec := NewItemWithPricingSpecification()
		assert.NoError(t, spec.IsSatisfiedBy(&item))
	})
}
