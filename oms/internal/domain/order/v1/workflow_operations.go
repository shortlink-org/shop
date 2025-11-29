package v1

// Workflow operation names for Temporal workflows.
// These are used for signals and queries, not domain events.
const (
	// WorkflowSignalCancel is the signal name for cancelling an order
	WorkflowSignalCancel = "order.cancel"

	// WorkflowSignalComplete is the signal name for completing an order
	WorkflowSignalComplete = "order.complete"

	// WorkflowQueryGet is the query name for getting order state
	WorkflowQueryGet = "order.get"
)

