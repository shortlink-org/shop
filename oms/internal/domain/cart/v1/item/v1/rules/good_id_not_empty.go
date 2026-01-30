package rules

import (
	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// GoodIdNotEmptySpec validates that goodId is not empty.
type GoodIdNotEmptySpec struct{}

func (s GoodIdNotEmptySpec) IsSatisfiedBy(item *itemv1.Item) error {
	if item.GetGoodId() == uuid.Nil {
		return itemv1.ErrItemGoodIdZero
	}
	return nil
}
