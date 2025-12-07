package v1

import (
	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
)

// GetItems returns a copy of the cart items.
func (s *State) GetItems() itemsv1.Items {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy to prevent external modification and race conditions
	itemsCopy := make(itemsv1.Items, len(s.items))
	copy(itemsCopy, s.items)

	return itemsCopy
}

