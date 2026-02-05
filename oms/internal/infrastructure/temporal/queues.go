package temporal

import (
	v1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
)

// TaskQueueNames maps domain task queues to their Temporal queue names.
// This is where infrastructure-specific naming lives.
var TaskQueueNames = map[v1.TaskQueue]string{
	v1.CartTaskQueue:  "CART_TASK_QUEUE",
	v1.OrderTaskQueue: "ORDER_TASK_QUEUE",
}

// GetQueueName returns the Temporal queue name for a domain task queue.
func GetQueueName(queue v1.TaskQueue) string {
	if name, ok := TaskQueueNames[queue]; ok {
		return name
	}
	// Fallback to domain string representation if mapping not found
	return queue.String()
}
