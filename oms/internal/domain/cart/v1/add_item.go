package v1

import (
	"errors"
	"fmt"
	"time"

	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/events/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// ErrInvalidCartItem is returned when a cart item fails validation.
var ErrInvalidCartItem = errors.New("invalid cart item")

// AddItem adds an item to the cart.
// If the item already exists, it increments the quantity immutably.
func (s *State) AddItem(item itemv1.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate item before adding
	if !item.IsValid() {
		return fmt.Errorf("invalid cart item goodId=%s quantity=%d: %w", item.GetGoodId(), item.GetQuantity(), ErrInvalidCartItem)
	}

	// Check if the item already exists in the cart
	for i, cartItem := range s.items {
		if cartItem.GetGoodId() == item.GetGoodId() {
			// Create a new item with updated quantity (immutable update)
			updatedItem, err := cartItem.WithQuantity(cartItem.GetQuantity() + item.GetQuantity())
			if err != nil {
				return fmt.Errorf("failed to update item quantity: %w", err)
			}

			s.items[i] = updatedItem
			// Generate domain event for item added/updated
			s.addDomainEvent(&eventsv1.ItemAddedEvent{
				CustomerID: s.customerId,
				Item:       updatedItem,
				OccurredAt: time.Now(),
			})

			return nil
		}
	}

	// Item doesn't exist, add it
	s.items = append(s.items, item)
	// Generate domain event for item added
	s.addDomainEvent(&eventsv1.ItemAddedEvent{
		CustomerID: s.customerId,
		Item:       item,
		OccurredAt: time.Now(),
	})

	return nil
}
