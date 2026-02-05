package rules

import (
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// DiscountNonNegativeSpec validates that discount is not negative.
type DiscountNonNegativeSpec struct{}

func (s DiscountNonNegativeSpec) IsSatisfiedBy(item *itemv1.Item) error {
	if item.GetDiscount().IsNegative() {
		return itemv1.ErrItemDiscountNegative
	}

	return nil
}
