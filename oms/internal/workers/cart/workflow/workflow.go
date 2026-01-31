package cart_workflow

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	v2 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/workers/cart/activities"
)

// Workflow is a Temporal workflow for cart operations.
//
// NOTE: For simple cart CRUD operations, direct database access via CartRepository
// is preferred over using this workflow. This workflow is kept for scenarios where
// you need:
// - Long-running cart sessions with TTL
// - Complex cart validation workflows
// - Integration with external services (stock reservation, etc.)
//
// For most use cases, use the cart UseCase directly instead of this workflow.
func Workflow(ctx workflow.Context, customerID uuid.UUID) error {
	logger := workflow.GetLogger(ctx)

	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Signal channels
	addChannel := workflow.GetSignalChannel(ctx, v2.Event_EVENT_ADD.String())
	removeChannel := workflow.GetSignalChannel(ctx, v2.Event_EVENT_REMOVE.String())
	resetChannel := workflow.GetSignalChannel(ctx, v2.Event_EVENT_RESET.String())

	// Cart session timeout (e.g., 24 hours)
	sessionTimeout := workflow.NewTimer(ctx, 24*time.Hour)

	selector := workflow.NewSelector(ctx)

	// Handle add item signal
	selector.AddReceive(addChannel, func(c workflow.ReceiveChannel, _ bool) {
		var req activities.AddItemRequest
		c.Receive(ctx, &req)

		logger.Info("Adding item to cart via activity", "customerID", customerID, "goodID", req.GoodID)
		err := workflow.ExecuteActivity(ctx, "AddItem", req).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to add item", "error", err)
		}
	})

	// Handle remove item signal
	selector.AddReceive(removeChannel, func(c workflow.ReceiveChannel, _ bool) {
		var req activities.RemoveItemRequest
		c.Receive(ctx, &req)

		logger.Info("Removing item from cart via activity", "customerID", customerID, "goodID", req.GoodID)
		err := workflow.ExecuteActivity(ctx, "RemoveItem", req).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to remove item", "error", err)
		}
	})

	// Handle reset signal
	selector.AddReceive(resetChannel, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, nil)

		logger.Info("Resetting cart via activity", "customerID", customerID)
		err := workflow.ExecuteActivity(ctx, "ResetCart", activities.ResetCartRequest{
			CustomerID: customerID,
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to reset cart", "error", err)
		}
	})

	// Handle session timeout
	selector.AddFuture(sessionTimeout, func(f workflow.Future) {
		logger.Info("Cart session timed out, resetting cart", "customerID", customerID)
		_ = workflow.ExecuteActivity(ctx, "ResetCart", activities.ResetCartRequest{
			CustomerID: customerID,
		}).Get(ctx, nil)
	})

	// Process signals until timeout
	for {
		selector.Select(ctx)
	}
}
