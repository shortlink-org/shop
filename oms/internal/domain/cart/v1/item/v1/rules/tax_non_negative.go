package rules

import (
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// TaxNonNegativeSpec validates that tax is not negative.
type TaxNonNegativeSpec struct{}

func (s TaxNonNegativeSpec) IsSatisfiedBy(item *itemv1.Item) error {
	if item.GetTax().IsNegative() {
		return itemv1.ErrItemTaxNegative
	}

	return nil
}
