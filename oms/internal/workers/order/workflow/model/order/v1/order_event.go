package v1

// OrderEvent represents an order event in workflow.
// Note: This is a legacy structure. Use domain event messages (OrderCreated, OrderCancelled, etc.) instead.
type OrderEvent struct {
	// EventType is the type of event (e.g., "order.created", "order.cancelled")
	EventType string
	// Items are the order items
	Items []*WorkerOrderItem
}
