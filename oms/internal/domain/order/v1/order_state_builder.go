package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/shortlink-org/go-sdk/fsm"

	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
)

// Error definitions
var (
	ErrInvalidOrderID              = errors.New("invalid order id")
	ErrInvalidGoodID               = errors.New("invalid good id")
	ErrInvalidOrderStatus          = errors.New("invalid order status")
	ErrCannotTransitionToCompleted = errors.New("cannot transition to COMPLETED from current status")
	ErrUnsupportedTargetStatus     = errors.New("unsupported target status")
	ErrOrderMustHaveItems          = errors.New("order in state must have at least one item")
)

// OrderStateBuilder is used to build a new OrderState
type OrderStateBuilder struct {
	orderState *OrderState
	errors     error
}

// NewOrderStateBuilder returns a new instance of OrderStateBuilder
func NewOrderStateBuilder(customerId uuid.UUID) *OrderStateBuilder {
	return &OrderStateBuilder{orderState: NewOrderState(customerId)}
}

// SetId sets the id of the order
func (b *OrderStateBuilder) SetId(id uuid.UUID) *OrderStateBuilder {
	if id == uuid.Nil {
		b.errors = errors.Join(b.errors, ErrInvalidOrderID)
		return b
	}

	b.orderState.id = id

	return b
}

// AddItem adds an item to the order
func (b *OrderStateBuilder) AddItem(goodId uuid.UUID, quantity int32, price decimal.Decimal) *OrderStateBuilder {
	if goodId == uuid.Nil {
		b.errors = errors.Join(b.errors, ErrInvalidGoodID)
		return b
	}

	item := NewItem(goodId, quantity, price)

	// Validate item before adding to maintain invariants
	err := ValidateOrderItem(item)
	if err != nil {
		b.errors = errors.Join(b.errors, fmt.Errorf("invalid item %s: %w", goodId, err))
		return b
	}

	b.orderState.items = append(b.orderState.items, item)

	return b
}

// SetStatus sets the status of the order by playing back the sequence of events needed to reach the desired status.
// This ensures that the FSM transitions through all required states correctly.
// Domain layer should not depend on context.Context from application layer.
func (b *OrderStateBuilder) SetStatus(targetStatus OrderStatus) *OrderStateBuilder {
	if targetStatus == OrderStatus_ORDER_STATUS_UNSPECIFIED {
		b.errors = errors.Join(b.errors, ErrInvalidOrderStatus)
		return b
	}

	currentStatus := b.orderState.GetStatus()

	// If already at the target status, no transition needed
	if currentStatus == targetStatus {
		return b
	}

	// Play back the sequence of events needed to reach the target status
	err := b.replayEventsToStatus(currentStatus, targetStatus)
	if err != nil {
		b.errors = errors.Join(b.errors, err)
		return b
	}

	return b
}

// replayEventsToStatus replays the sequence of FSM events needed to transition from current to target status
// Domain layer uses context.Background() internally for FSM, keeping domain pure
func (b *OrderStateBuilder) replayEventsToStatus(currentStatus, targetStatus OrderStatus) error {
	// Define the transition path
	switch targetStatus {
	case OrderStatus_ORDER_STATUS_PROCESSING:
		// PENDING + CREATE => PROCESSING
		if currentStatus == OrderStatus_ORDER_STATUS_PENDING {
			err := b.orderState.fsm.TriggerEvent(context.Background(), fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CREATE.String()))
			if err != nil {
				return fmt.Errorf("failed to transition to PROCESSING: %w", err)
			}
		}

	case OrderStatus_ORDER_STATUS_CANCELLED:
		// PENDING/PROCESSING + CANCEL => CANCELLED
		err := b.orderState.fsm.TriggerEvent(context.Background(), fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CANCEL.String()))
		if err != nil {
			return fmt.Errorf("failed to transition to CANCELLED: %w", err)
		}

	case OrderStatus_ORDER_STATUS_COMPLETED:
		// PENDING: first CREATE => PROCESSING, then COMPLETE => COMPLETED
		if currentStatus == OrderStatus_ORDER_STATUS_PENDING {
			err := b.orderState.fsm.TriggerEvent(context.Background(), fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_CREATE.String()))
			if err != nil {
				return fmt.Errorf("failed to transition to PROCESSING: %w", err)
			}

			currentStatus = OrderStatus_ORDER_STATUS_PROCESSING
		}

		if currentStatus == OrderStatus_ORDER_STATUS_PROCESSING {
			err := b.orderState.fsm.TriggerEvent(context.Background(), fsm.Event(commonv1.OrderTransitionEvent_ORDER_TRANSITION_EVENT_COMPLETE.String()))
			if err != nil {
				return fmt.Errorf("failed to transition to COMPLETED: %w", err)
			}
		} else {
			return fmt.Errorf("cannot transition to COMPLETED from status %s: %w", currentStatus, ErrCannotTransitionToCompleted)
		}

	default:
		return fmt.Errorf("unsupported target status %s: %w", targetStatus, ErrUnsupportedTargetStatus)
	}

	return nil
}

// Build finalizes the building process and returns the built OrderState
func (b *OrderStateBuilder) Build() (*OrderState, error) {
	if b.errors != nil {
		return nil, b.errors
	}

	// Validate order invariants before returning
	err := ValidateOrderItems(b.orderState.items)
	if err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Validate that order has items if status is PROCESSING or COMPLETED
	currentStatus := b.orderState.GetStatus()
	if currentStatus == OrderStatus_ORDER_STATUS_PROCESSING || currentStatus == OrderStatus_ORDER_STATUS_COMPLETED {
		if len(b.orderState.items) == 0 {
			return nil, fmt.Errorf("order in %s state must have at least one item: %w", currentStatus, ErrOrderMustHaveItems)
		}
	}

	return b.orderState, nil
}
