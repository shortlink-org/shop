package rules

import (
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// QuantityPositiveSpec validates that quantity is greater than zero.
type QuantityPositiveSpec struct{}

func (s QuantityPositiveSpec) IsSatisfiedBy(item *itemv1.Item) error {
	if item.GetQuantity() <= 0 {
		return itemv1.ErrItemQuantityZero
	}

	return nil
}
