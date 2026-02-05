package rules

import (
	"fmt"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// DiscountNotExceedsPriceSpec validates that discount does not exceed price.
type DiscountNotExceedsPriceSpec struct{}

func (s DiscountNotExceedsPriceSpec) IsSatisfiedBy(item *itemv1.Item) error {
	if item.GetDiscount().GreaterThan(item.GetPrice()) {
		return fmt.Errorf("%w: discount %s exceeds price %s",
			itemv1.ErrItemDiscountExceedsPrice, item.GetDiscount().String(), item.GetPrice().String())
	}

	return nil
}
