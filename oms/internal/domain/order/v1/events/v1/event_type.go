package v1

// EventType returns the canonical event type for subscription/routing.
func (*OrderCreated) EventType() string { return "oms.order.created.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderCancelled) EventType() string { return "oms.order.cancelled.v1" }

// EventType returns the canonical event type for subscription/routing.
func (*OrderCompleted) EventType() string { return "oms.order.completed.v1" }
