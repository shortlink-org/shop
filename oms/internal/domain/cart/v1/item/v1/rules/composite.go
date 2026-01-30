package rules

import (
	"github.com/shortlink-org/go-sdk/specification"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// NewItemSpecification returns a composite specification for basic Item validation.
// Validates: goodId not empty, quantity > 0.
func NewItemSpecification() specification.Specification[itemv1.Item] {
	return specification.NewAndSpecification[itemv1.Item](
		GoodIdNotEmptySpec{},
		QuantityPositiveSpec{},
	)
}

// NewItemWithPricingSpecification returns a composite specification for full Item validation.
// Validates: goodId not empty, quantity > 0, price >= 0, discount >= 0, tax >= 0, discount <= price.
func NewItemWithPricingSpecification() specification.Specification[itemv1.Item] {
	return specification.NewAndSpecification[itemv1.Item](
		GoodIdNotEmptySpec{},
		QuantityPositiveSpec{},
		PriceNonNegativeSpec{},
		DiscountNonNegativeSpec{},
		TaxNonNegativeSpec{},
		DiscountNotExceedsPriceSpec{},
	)
}
