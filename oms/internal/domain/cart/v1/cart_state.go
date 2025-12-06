package v1

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CartState represents the cart state.
type CartState struct {
	mu sync.Mutex

	// items is the cart items
	items CartItems
	// customerId is the customer ID
	customerId uuid.UUID
	// domainEvents stores domain events that occurred during aggregate operations
	domainEvents []DomainEvent
}

// NewCartState creates a new cart state.
func NewCartState(customerId uuid.UUID) *CartState {
	return &CartState{
		items:        make([]CartItem, 0),
		customerId:   customerId,
		domainEvents: make([]DomainEvent, 0),
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
				// Generate domain event for item added/updated
				m.addDomainEvent(&CartItemAddedEvent{
					CustomerID: m.customerId,
					Item:       updatedItem,
					OccurredAt: time.Now(),
				})
				return nil
			}
		}

		// Item doesn't exist, add it
		m.items = append(m.items, item)
		// Generate domain event for item added
		m.addDomainEvent(&CartItemAddedEvent{
			CustomerID: m.customerId,
			Item:       item,
			OccurredAt: time.Now(),
		})
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
					// Generate domain event for item removed
					m.addDomainEvent(&CartItemRemovedEvent{
						CustomerID: m.customerId,
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
				m.items[i] = updatedItem
				// Generate domain event for item removed (quantity decreased)
				m.addDomainEvent(&CartItemRemovedEvent{
					CustomerID: m.customerId,
					Item:       item, // Item being removed
					OccurredAt: time.Now(),
				})
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
	// Generate domain event for cart reset
	m.addDomainEvent(&CartResetEvent{
		CustomerID: m.customerId,
		OccurredAt: time.Now(),
	})
}

// addDomainEvent adds a domain event to the aggregate's event list
func (m *CartState) addDomainEvent(event DomainEvent) {
	m.domainEvents = append(m.domainEvents, event)
}

// GetDomainEvents returns all domain events that occurred during aggregate operations
// Application layer should call this after aggregate operations to publish events
// and then call ClearDomainEvents() to reset the list
func (m *CartState) GetDomainEvents() []DomainEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Return a copy to prevent external modification
	eventsCopy := make([]DomainEvent, len(m.domainEvents))
	copy(eventsCopy, m.domainEvents)
	return eventsCopy
}

// ClearDomainEvents clears the domain events list
// Should be called by application layer after publishing events
func (m *CartState) ClearDomainEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.domainEvents = m.domainEvents[:0]
}
