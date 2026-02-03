package v1

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/shortlink-org/go-sdk/fsm"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common/v1"
	eventsv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/events/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	// domainEvents stores domain events (proto) that occurred during aggregate operations
	domainEvents []any
	// deliveryInfo contains delivery information for the order (nil = self-pickup)
	deliveryInfo *DeliveryInfo
	// deliveryStatus tracks the delivery status (ACCEPTED, ASSIGNED, IN_TRANSIT, etc.)
	deliveryStatus commonv1.DeliveryStatus
}

// NewOrderState creates a new OrderState instance with the given customer ID.
func NewOrderState(customerId uuid.UUID) *OrderState {
	return newOrderState(uuid.New(), customerId, make(Items, 0), OrderStatus_ORDER_STATUS_PENDING, 0, nil)
}

// NewOrderStateFromPersisted builds an OrderState from persisted data (repository load).
// Single constructor for both "new order" and "reconstitute"; FSM rules live only here.
func NewOrderStateFromPersisted(id uuid.UUID, customerId uuid.UUID, items Items, status OrderStatus, version int, deliveryInfo *DeliveryInfo) *OrderState {
	if items == nil {
		items = make(Items, 0)
	}
	return newOrderState(id, customerId, items, status, version, deliveryInfo)
}

// newOrderState is the single place that builds OrderState and configures the FSM.
func newOrderState(id uuid.UUID, customerId uuid.UUID, items Items, status OrderStatus, version int, deliveryInfo *DeliveryInfo) *OrderState {
	order := &OrderState{
		id:           id,
		items:        items,
		customerId:   customerId,
		version:      version,
		domainEvents: make([]any, 0),
		deliveryInfo: deliveryInfo,
	}
	order.fsm = fsm.New(fsm.State(status.String()))
	order.addOrderTransitionRules(order.fsm)
	order.fsm.SetOnEnterState(order.onEnterState)
	order.fsm.SetOnExitState(order.onExitState)
	return order
}

// addOrderTransitionRules registers the order FSM transition rules (single source of truth).
func (o *OrderState) addOrderTransitionRules(f *fsm.FSM) {
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELLED.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELLED.String()),
	)
	f.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_COMPLETED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_COMPLETED.String()),
	)
}

// GetVersion returns the current version for optimistic concurrency control.
func (o *OrderState) GetVersion() int {
	return o.version
}

// SetID sets the order ID (used when persisting a new order).
func (o *OrderState) SetID(id uuid.UUID) {
	o.id = id
}

// onEnterState is the callback executed when entering a new state.
// This is where we generate domain events (proto) based on state transitions.
func (o *OrderState) onEnterState(ctx context.Context, from, to fsm.State, event fsm.Event) {
	toStatus := o.fsmStateToOrderStatus(to)
	now := time.Now()
	ts := timestamppb.New(now)

	switch toStatus {
	case OrderStatus_ORDER_STATUS_PROCESSING:
		if o.fsmStateToOrderStatus(from) == OrderStatus_ORDER_STATUS_PENDING {
			items := make([]*commonv1.OrderItem, 0, len(o.items))
			for _, it := range o.items {
				items = append(items, &commonv1.OrderItem{
					GoodId:   it.GetGoodId().String(),
					Quantity: it.GetQuantity(),
					Price:    it.GetPrice().String(),
				})
			}
			o.addDomainEvent(&eventsv1.OrderCreated{
				OrderId:    o.id.String(),
				CustomerId: o.customerId.String(),
				Items:      items,
				Status:     toStatus,
				CreatedAt:  ts,
				OccurredAt: ts,
			})
		}
	case OrderStatus_ORDER_STATUS_CANCELLED:
		o.addDomainEvent(&eventsv1.OrderCancelled{
			OrderId:     o.id.String(),
			CustomerId:  o.customerId.String(),
			Status:      toStatus,
			Reason:      "", // Set by command layer when CancelOrder(reason) is added
			CancelledAt: ts,
			OccurredAt:  ts,
		})
	case OrderStatus_ORDER_STATUS_COMPLETED:
		o.addDomainEvent(&eventsv1.OrderCompleted{
			OrderId:     o.id.String(),
			CustomerId:  o.customerId.String(),
			Status:      toStatus,
			CompletedAt: ts,
			OccurredAt:  ts,
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
	if o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED ||
		o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT ||
		o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED ||
		o.deliveryStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED {
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
func (o *OrderState) GetDeliveryStatus() commonv1.DeliveryStatus {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.deliveryStatus
}

// SetDeliveryStatus updates the delivery status.
// Returns an error if the order is in a terminal state or if the transition is invalid.
func (o *OrderState) SetDeliveryStatus(status commonv1.DeliveryStatus) error {
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
func (o *OrderState) isValidDeliveryStatusTransition(from, to commonv1.DeliveryStatus) bool {
	// Allow setting initial status
	if from == commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED {
		return true
	}

	// Define valid transitions
	validTransitions := map[commonv1.DeliveryStatus][]commonv1.DeliveryStatus{
		commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED: {
			commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED,
		},
		commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED: {
			commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT,
		},
		commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT: {
			commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED,
			commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED,
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

// addDomainEvent adds a domain event (proto) to the aggregate's event list.
func (o *OrderState) addDomainEvent(event any) {
	o.domainEvents = append(o.domainEvents, event)
}

// GetDomainEvents returns all domain events that occurred during aggregate operations.
// Application layer should publish them (e.g. to outbox) then call ClearDomainEvents().
func (o *OrderState) GetDomainEvents() []any {
	o.mu.Lock()
	defer o.mu.Unlock()

	eventsCopy := make([]any, len(o.domainEvents))
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
