package v1

import (
	"time"

	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// ItemRemovedEvent represents the domain event when an item is removed from the cart
type ItemRemovedEvent struct {
	CustomerID uuid.UUID
	Item       itemv1.Item
	OccurredAt time.Time
}

func (e *ItemRemovedEvent) EventType() string {
	return "ItemRemoved"
}
