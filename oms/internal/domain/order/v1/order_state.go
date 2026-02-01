package v1

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/shortlink-org/go-sdk/fsm"
	common "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common/v1"
)

// OrderState represents the order state.
type OrderState struct {
	mu sync.Mutex

	// orderID is the order ID
	id uuid.UUID
	// items is the list of order items
	items Items
	// customerId is the customer ID
	customerId uuid.UUID
	// version is used for optimistic concurrency control
	version int
	// fsm is the finite state machine for the order status
	fsm *fsm.FSM
	// domainEvents stores domain events that occurred during aggregate operations
	// Events are collected here and can be retrieved by application layer for publishing
	domainEvents []DomainEvent
	// deliveryInfo contains delivery information for the order (nil = self-pickup)
	deliveryInfo *DeliveryInfo
	// deliveryStatus tracks the delivery status (ACCEPTED, ASSIGNED, IN_TRANSIT, etc.)
	deliveryStatus common.DeliveryStatus
}

// NewOrderState creates a new OrderState instance with the given customer ID.
func NewOrderState(customerId uuid.UUID) *OrderState {
	order := &OrderState{
		id:           uuid.New(),
		items:        make(Items, 0),
		customerId:   customerId,
		version:      0,
		domainEvents: make([]DomainEvent, 0),
	}

	// Initialize the FSM with the initial state.
	order.fsm = fsm.New(fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()))

	// Define transition rules.
	order.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	order.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	order.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELLED.String()),
	)
	order.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELLED.String()),
	)
	order.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_COMPLETED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_COMPLETED.String()),
	)

	// Set up callbacks.
	order.fsm.SetOnEnterState(order.onEnterState)
	order.fsm.SetOnExitState(order.onExitState)

	return order
}

// onEnterState is the callback executed when entering a new state.
// This is where we generate domain events based on state transitions.
func (o *OrderState) onEnterState(ctx context.Context, from, to fsm.State, event fsm.Event) {
	// Convert FSM state to OrderStatus
	toStatus := o.fsmStateToOrderStatus(to)

	// Generate domain events based on state transitions
	now := time.Now()
	switch toStatus {
	case OrderStatus_ORDER_STATUS_PROCESSING:
		// OrderCreated event - when transitioning to PROCESSING from PENDING
		if o.fsmStateToOrderStatus(from) == OrderStatus_ORDER_STATUS_PENDING {
			// Create a copy of items for the event
			itemsCopy := make(Items, len(o.items))
			copy(itemsCopy, o.items)
			o.addDomainEvent(&OrderCreatedEvent{
				OrderID:    o.id,
				CustomerID: o.customerId,
				Items:      itemsCopy,
				Status:     toStatus,
				OccurredAt: now,
			})
		}
	case OrderStatus_ORDER_STATUS_CANCELLED:
		// OrderCancelled event
		o.addDomainEvent(&OrderCancelledEvent{
			OrderID:    o.id,
			CustomerID: o.customerId,
			Status:     toStatus,
			OccurredAt: now,
		})
	case OrderStatus_ORDER_STATUS_COMPLETED:
		// OrderCompleted event
		o.addDomainEvent(&OrderCompletedEvent{
			OrderID:    o.id,
			CustomerID: o.customerId,
			Status:     toStatus,
			OccurredAt: now,
		})
	}
}

// onExitState is the callback executed when exiting a state.
// Domain layer should not depend on infrastructure (logging, stdout, etc.).
func (o *OrderState) onExitState(ctx context.Context, from, to fsm.State, event fsm.Event) {
	// Domain layer should not perform side effects like logging.
	// Any logging/observability should be implemented at the application/infrastructure layer.
}

// GetOrderID returns the unique identifier of the order.
func (o *OrderState) GetOrderID() uuid.UUID {
	return o.id
}

// GetItems returns a copy of the list of items in the order.
func (o *OrderState) GetItems() Items {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Return a copy to prevent external modification.
	itemsCopy := make(Items, len(o.items))
	copy(itemsCopy, o.items)

	return itemsCopy
}

// GetCustomerId returns the customer ID associated with the order.
func (o *OrderState) GetCustomerId() uuid.UUID {
	return o.customerId
}

// GetDeliveryInfo returns the delivery information for the order.
// Returns nil if the order is for self-pickup (no delivery).
func (o *OrderState) GetDeliveryInfo() *DeliveryInfo {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryInfo
}

// SetDeliveryInfo sets the delivery information for the order.
// Returns an error if the order is in a terminal state, delivery is already in progress,
// or if the delivery info is invalid.
func (o *OrderState) SetDeliveryInfo(info DeliveryInfo) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Validate delivery info (invariant: delivery info must be valid)
	if !info.IsValid() {
		return fmt.Errorf("invalid delivery info: address, delivery period and package info are required")
	}

	// Check OrderStatus - cannot update delivery info in terminal states
	currentStatus := o.getStatusUnlocked()
	if currentStatus == OrderStatus_ORDER_STATUS_COMPLETED ||
		currentStatus == OrderStatus_ORDER_STATUS_CANCELLED {
		return fmt.Errorf("cannot update delivery info in %s state", currentStatus)
	}

	// Check DeliveryStatus - cannot change after package is assigned to courier or in transit
	if o.deliveryStatus == common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED ||
		o.deliveryStatus == common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT ||
		o.deliveryStatus == common.DeliveryStatus_DELIVERY_STATUS_DELIVERED ||
		o.deliveryStatus == common.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED {
		return fmt.Errorf("cannot update delivery info: package already %s", o.deliveryStatus)
	}

	o.deliveryInfo = &info
	return nil
}

// HasDeliveryInfo returns true if the order has delivery information.
func (o *OrderState) HasDeliveryInfo() bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryInfo != nil
}

// GetDeliveryStatus returns the current delivery status.
func (o *OrderState) GetDeliveryStatus() common.DeliveryStatus {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryStatus
}

// SetDeliveryStatus updates the delivery status.
// Returns an error if the order is in a terminal state or if the transition is invalid.
func (o *OrderState) SetDeliveryStatus(status common.DeliveryStatus) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Cannot update delivery status in terminal order states
	currentOrderStatus := o.getStatusUnlocked()
	if currentOrderStatus == OrderStatus_ORDER_STATUS_COMPLETED ||
		currentOrderStatus == OrderStatus_ORDER_STATUS_CANCELLED {
		return fmt.Errorf("cannot update delivery status in %s order state", currentOrderStatus)
	}

	// Validate delivery status transition (only forward transitions allowed)
	if !o.isValidDeliveryStatusTransition(o.deliveryStatus, status) {
		return fmt.Errorf("invalid delivery status transition from %s to %s", o.deliveryStatus, status)
	}

	o.deliveryStatus = status
	return nil
}

// isValidDeliveryStatusTransition checks if the delivery status transition is valid.
// Delivery status can only move forward: UNSPECIFIED -> ACCEPTED -> ASSIGNED -> IN_TRANSIT -> DELIVERED/NOT_DELIVERED
func (o *OrderState) isValidDeliveryStatusTransition(from, to common.DeliveryStatus) bool {
	// Allow setting initial status
	if from == common.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED {
		return true
	}

	// Define valid transitions
	validTransitions := map[common.DeliveryStatus][]common.DeliveryStatus{
		common.DeliveryStatus_DELIVERY_STATUS_ACCEPTED: {
			common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
		},
		common.DeliveryStatus_DELIVERY_STATUS_ASSIGNED: {
			common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
		},
		common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT: {
			common.DeliveryStatus_DELIVERY_STATUS_DELIVERED,
			common.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED,
		},
	}

	allowedTargets, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if to == allowed {
			return true
		}
	}

	return false
}

// GetStatus returns the current status of the order.
func (o *OrderState) GetStatus() OrderStatus {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.getStatusUnlocked()
}

// getStatusUnlocked returns the current status without locking (for internal use).
func (o *OrderState) getStatusUnlocked() OrderStatus {
	return o.fsmStateToOrderStatus(o.fsm.GetCurrentState())
}

// fsmStateToOrderStatus converts FSM state to OrderStatus enum
func (o *OrderState) fsmStateToOrderStatus(state fsm.State) OrderStatus {
	for k, v := range OrderStatus_name {
		if v == state.String() {
			return OrderStatus(k)
		}
	}
	return OrderStatus_ORDER_STATUS_UNSPECIFIED
}

// addDomainEvent adds a domain event to the aggregate's event list
func (o *OrderState) addDomainEvent(event DomainEvent) {
	o.domainEvents = append(o.domainEvents, event)
}

// GetDomainEvents returns all domain events that occurred during aggregate operations
// Application layer should call this after aggregate operations to publish events
// and then call ClearDomainEvents() to reset the list
func (o *OrderState) GetDomainEvents() []DomainEvent {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Return a copy to prevent external modification
	eventsCopy := make([]DomainEvent, len(o.domainEvents))
	copy(eventsCopy, o.domainEvents)
	return eventsCopy
}

// ClearDomainEvents clears the domain events list
// Should be called by application layer after publishing events
func (o *OrderState) ClearDomainEvents() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.domainEvents = o.domainEvents[:0]
}

// CreateOrder initializes the order with the provided items and transitions it to Processing state.
// Domain layer should not depend on context.Context from application layer.
// We use context.Background() internally for FSM, keeping domain pure.
func (o *OrderState) CreateOrder(items Items) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Validate items before creating order (invariant: order must have valid items)
	if err := ValidateOrderItems(items); err != nil {
		return fmt.Errorf("cannot create order: %w", err)
	}

	// Create a defensive copy to prevent external modification
	itemsCopy := make(Items, len(items))
	copy(itemsCopy, items)

	// Validate state transition before triggering FSM event
	// (invariant: items must be valid before transitioning to PROCESSING)
	currentStatus := o.getStatusUnlocked()
	if err := ValidateOrderStateTransition(currentStatus, OrderStatus_ORDER_STATUS_PROCESSING, itemsCopy); err != nil {
		return err
	}

	// Trigger the transition event to Processing.
	// Use context.Background() to keep domain layer independent of application context
	err := o.fsm.TriggerEvent(context.Background(), fsm.Event(OrderStatus_ORDER_STATUS_PENDING.String()))
	if err != nil {
		return err
	}

	// Set items after successful transition (invariant: items are set only after validation)
	o.items = itemsCopy
	return nil
}

// UpdateOrder updates the order's items. It modifies existing items and adds new ones as needed.
// Domain layer should not depend on context.Context from application layer.
func (o *OrderState) UpdateOrder(items Items) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Cannot update order in terminal states
	currentStatus := o.getStatusUnlocked()
	if currentStatus == OrderStatus_ORDER_STATUS_COMPLETED || currentStatus == OrderStatus_ORDER_STATUS_CANCELLED {
		return fmt.Errorf("cannot update order in %s state", currentStatus)
	}

	// Create a map for quick lookup of existing item indices.
	itemIndexMap := make(map[uuid.UUID]int)
	for i := range o.items {
		itemIndexMap[o.items[i].goodId] = i
	}

	// Create updated items list
	updatedItems := make(Items, 0, len(o.items)+len(items))

	// First, update existing items
	for i := range o.items {
		updatedItems = append(updatedItems, o.items[i])
	}

	// Update quantities and prices of existing items or add new items.
	for _, item := range items {
		if idx, exists := itemIndexMap[item.goodId]; exists {
			// Validate updated item
			if err := ValidateOrderItem(item); err != nil {
				return fmt.Errorf("cannot update item %s: %w", item.goodId, err)
			}
			// Update existing item by index
			updatedItems[idx].quantity = item.quantity
			updatedItems[idx].price = item.price
		} else {
			// Validate new item before adding
			if err := ValidateOrderItem(item); err != nil {
				return fmt.Errorf("cannot add item %s: %w", item.goodId, err)
			}
			// Append new items
			updatedItems = append(updatedItems, item)
		}
	}

	// Validate the resulting items list (invariant: order must have valid items)
	if err := ValidateOrderItems(updatedItems); err != nil {
		return fmt.Errorf("cannot update order: %w", err)
	}

	// Set updated items
	o.items = updatedItems
	return nil
}

// CancelOrder transitions the order to the Cancelled state.
// Domain layer should not depend on context.Context from application layer.
func (o *OrderState) CancelOrder() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Trigger the transition event to Cancel.
	// Use context.Background() to keep domain layer independent of application context
	err := o.fsm.TriggerEvent(context.Background(), fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()))
	if err != nil {
		return err
	}

	return nil
}

// CompleteOrder transitions the order to the Completed state.
// Domain layer should not depend on context.Context from application layer.
func (o *OrderState) CompleteOrder() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	currentStatus := o.getStatusUnlocked()

	// Validate state transition (invariant: can only complete orders in PROCESSING state)
	if currentStatus != OrderStatus_ORDER_STATUS_PROCESSING {
		return fmt.Errorf("cannot complete order in %s state, must be PROCESSING", currentStatus)
	}

	// Validate items before completing (invariant: order must have valid items to complete)
	if err := ValidateOrderItems(o.items); err != nil {
		return fmt.Errorf("cannot complete order with invalid items: %w", err)
	}

	// Trigger the transition event to Complete.
	// Use context.Background() to keep domain layer independent of application context
	err := o.fsm.TriggerEvent(context.Background(), fsm.Event(OrderStatus_ORDER_STATUS_COMPLETED.String()))
	if err != nil {
		return err
	}

	return nil
}
