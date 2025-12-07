package v1

import (
	"fmt"
	"time"

	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/events/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
)

// RemoveItem removes an item from the cart.
// If the item exists, it decrements the quantity immutably.
func (s *State) RemoveItem(item itemv1.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the item in the cart
	for i, cartItem := range s.items {
		if cartItem.GetGoodId() == item.GetGoodId() {
			newQuantity := cartItem.GetQuantity() - item.GetQuantity()
			if newQuantity <= 0 {
				// Remove the item completely
				s.items = append(s.items[:i], s.items[i+1:]...)
				// Generate domain event for item removed
				s.addDomainEvent(&eventsv1.ItemRemovedEvent{
					CustomerID: s.customerId,
					Item:       cartItem, // Use original item before removal
					OccurredAt: time.Now(),
				})
				return nil
			}

			// Create a new item with updated quantity (immutable update)
			updatedItem, err := cartItem.WithQuantity(newQuantity)
			if err != nil {
				return fmt.Errorf("failed to update item quantity: %w", err)
			}
			s.items[i] = updatedItem
			// Generate domain event for item removed (quantity decreased)
			s.addDomainEvent(&eventsv1.ItemRemovedEvent{
				CustomerID: s.customerId,
				Item:       item, // Item being removed
				OccurredAt: time.Now(),
			})
			return nil
		}
	}

	// Item not found, nothing to remove
	return nil
}

