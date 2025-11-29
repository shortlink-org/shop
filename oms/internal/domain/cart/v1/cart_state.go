package v1

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// CartState represents the cart state.
type CartState struct {
	mu sync.Mutex

	// items is the cart items
	items CartItems
	// customerId is the customer ID
	customerId uuid.UUID
}

// NewCartState creates a new cart state.
func NewCartState(customerId uuid.UUID) *CartState {
	return &CartState{
		items:      make([]CartItem, 0),
		customerId: customerId,
	}
}

// GetItems returns a copy of the cart items.
func (m *CartState) GetItems() CartItems {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to prevent external modification and race conditions
	itemsCopy := make(CartItems, len(m.items))
	copy(itemsCopy, m.items)

	return itemsCopy
}

// GetCustomerId returns the value of the customerId field.
func (m *CartState) GetCustomerId() uuid.UUID {
	return m.customerId
}

// AddItem adds an item to the cart.
// If the item already exists, it increments the quantity immutably.
func (m *CartState) AddItem(item CartItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate item before adding
	if !item.IsValid() {
		return fmt.Errorf("invalid cart item: goodId=%s, quantity=%d", item.GetGoodId(), item.GetQuantity())
	}

	// Check if the item already exists in the cart
	for i, cartItem := range m.items {
		if cartItem.GetGoodId() == item.GetGoodId() {
			// Create a new item with updated quantity (immutable update)
			updatedItem, err := cartItem.WithQuantity(cartItem.GetQuantity() + item.GetQuantity())
			if err != nil {
				return fmt.Errorf("failed to update item quantity: %w", err)
			}
			m.items[i] = updatedItem
			return nil
		}
	}

	// Item doesn't exist, add it
	m.items = append(m.items, item)
	return nil
}

// RemoveItem removes an item from the cart.
// If the item exists, it decrements the quantity immutably.
func (m *CartState) RemoveItem(item CartItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the item in the cart
	for i, cartItem := range m.items {
		if cartItem.GetGoodId() == item.GetGoodId() {
			newQuantity := cartItem.GetQuantity() - item.GetQuantity()
			if newQuantity <= 0 {
				// Remove the item completely
				m.items = append(m.items[:i], m.items[i+1:]...)
				return nil
			}
			
			// Create a new item with updated quantity (immutable update)
			updatedItem, err := cartItem.WithQuantity(newQuantity)
			if err != nil {
				return fmt.Errorf("failed to update item quantity: %w", err)
			}
			m.items[i] = updatedItem
			return nil
		}
	}

	// Item not found, nothing to remove
	return nil
}

// Reset resets the cart.
func (m *CartState) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make([]CartItem, 0)
}
