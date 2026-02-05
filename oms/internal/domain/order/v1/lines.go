package v1

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Line is a neutral input for creating order items (e.g. from cart or API).
// Domain does not depend on cart or other application types.
type Line struct {
	ProductID uuid.UUID
	Qty       int32
	UnitPrice decimal.Decimal
}

// CreateFromLines initializes the order with the provided lines and transitions it to Processing state.
func (o *OrderState) CreateFromLines(ctx context.Context, lines []Line) error {
	items := make(Items, 0, len(lines))
	for _, l := range lines {
		items = append(items, NewItem(l.ProductID, l.Qty, l.UnitPrice))
	}

	return o.CreateOrder(ctx, items)
}
