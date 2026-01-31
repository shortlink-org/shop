package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/shortlink-org/go-sdk/fsm"
)

// Reconstitute creates an OrderState from persisted data.
// This is used by the repository to rebuild the aggregate from the database.
// It bypasses validation since the data is already validated when it was saved.
func Reconstitute(id uuid.UUID, customerId uuid.UUID, items Items, status OrderStatus, version int) *OrderState {
	order := &OrderState{
		id:           id,
		items:        items,
		customerId:   customerId,
		version:      version,
		domainEvents: make([]DomainEvent, 0),
	}

	// Initialize FSM with the persisted state
	order.fsm = fsm.New(fsm.State(status.String()))

	// Define transition rules (same as NewOrderState)
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

	// Set up callbacks
	order.fsm.SetOnEnterState(order.onEnterState)
	order.fsm.SetOnExitState(order.onExitState)

	return order
}

// GetVersion returns the current version for optimistic concurrency control.
func (o *OrderState) GetVersion() int {
	return o.version
}

// SetID sets the order ID (used during reconstitution).
func (o *OrderState) SetID(id uuid.UUID) {
	o.id = id
}

// initFSMWithStatus initializes the FSM with a given status.
// This is a helper for reconstitution to avoid duplicating FSM setup logic.
func (o *OrderState) initFSMWithStatus(ctx context.Context, status OrderStatus) {
	o.fsm = fsm.New(fsm.State(status.String()))

	// Define transition rules
	o.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	o.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
	)
	o.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PENDING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELLED.String()),
	)
	o.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_CANCELLED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_CANCELLED.String()),
	)
	o.fsm.AddTransitionRule(
		fsm.State(OrderStatus_ORDER_STATUS_PROCESSING.String()),
		fsm.Event(OrderStatus_ORDER_STATUS_COMPLETED.String()),
		fsm.State(OrderStatus_ORDER_STATUS_COMPLETED.String()),
	)

	// Set up callbacks
	o.fsm.SetOnEnterState(o.onEnterState)
	o.fsm.SetOnExitState(o.onExitState)
}
