package v1

// EventType returns the canonical event type for subscription/routing.
func (*OrderCreated) EventType() string { return "oms.order.created.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderCancelled) EventType() string { return "oms.order.cancelled.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderCompleted) EventType() string { return "oms.order.completed.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderDeliveryRequestedEvent) EventType() string { return "oms.order.delivery_requested.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderDeliveryStatusUpdatedEvent) EventType() string {
	return "oms.order.delivery_status_updated.v1"
}

// EventType returns the canonical event type for subscription/routing.
func (*OrderDeliveryCompletedEvent) EventType() string { return "oms.order.delivery_completed.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderDeliveryFailedEvent) EventType() string { return "oms.order.delivery_failed.v1" }
