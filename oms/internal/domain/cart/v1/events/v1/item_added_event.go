package v1

import (
	"time"

	"github.com/google/uuid"

	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// ItemAddedEvent represents the domain event when an item is added to the cart
type ItemAddedEvent struct {
	CustomerID uuid.UUID
	Item       itemv1.Item
	OccurredAt time.Time
}

func (e *ItemAddedEvent) EventType() string {
	return "ItemAdded"
}
