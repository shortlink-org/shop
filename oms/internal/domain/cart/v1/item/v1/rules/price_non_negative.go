package rules

import (
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// PriceNonNegativeSpec validates that price is not negative.
type PriceNonNegativeSpec struct{}

func (s PriceNonNegativeSpec) IsSatisfiedBy(item *itemv1.Item) error {
	if item.GetPrice().IsNegative() {
		return itemv1.ErrItemPriceNegative
	}
	return nil
}
