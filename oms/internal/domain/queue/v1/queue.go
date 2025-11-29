package v1

// TaskQueue represents a domain abstraction for task queues.
// It defines what queues exist in the domain, not how they are implemented.
type TaskQueue int

const (
	// CartTaskQueue represents the task queue for cart operations
	CartTaskQueue TaskQueue = iota
	// OrderTaskQueue represents the task queue for order operations
	OrderTaskQueue
)

// String returns a human-readable name for the task queue.
// This is for logging/debugging purposes only, not for infrastructure mapping.
func (tq TaskQueue) String() string {
	switch tq {
	case CartTaskQueue:
		return "cart"
	case OrderTaskQueue:
		return "order"
	default:
		return "unknown"
	}
}
