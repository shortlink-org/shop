package cart_workflow

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	v2 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
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
		StartToCloseTimeout: 10 * time.Second, //nolint:mnd // activity timeout
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0, //nolint:mnd // exponential backoff
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				itemv1.ErrItemGoodIdZero.Error(),
				itemv1.ErrItemQuantityZero.Error(),
				itemv1.ErrItemPriceNegative.Error(),
				itemv1.ErrItemDiscountNegative.Error(),
				itemv1.ErrItemTaxNegative.Error(),
				itemv1.ErrItemDiscountExceedsPrice.Error(),
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Signal channels
	addChannel := workflow.GetSignalChannel(ctx, v2.Event_EVENT_ADD.String())
	removeChannel := workflow.GetSignalChannel(ctx, v2.Event_EVENT_REMOVE.String())
	resetChannel := workflow.GetSignalChannel(ctx, v2.Event_EVENT_RESET.String())

	var (
		sessionTimeout       workflow.Future
		sessionTimeoutCtx    workflow.Context
		cancelSessionTimeout workflow.CancelFunc
	)

	resetSessionTimeout := func() {
		if cancelSessionTimeout != nil {
			cancelSessionTimeout()
		}

		sessionTimeoutCtx, cancelSessionTimeout = workflow.WithCancel(ctx)
		sessionTimeout = workflow.NewTimerWithOptions(sessionTimeoutCtx, 24*time.Hour, workflow.TimerOptions{ //nolint:mnd
			Summary: "Cart session TTL - auto-reset after 24 hours of inactivity",
		})
	}

	resetSessionTimeout()

	// Process signals until timeout
	for {
		selector := workflow.NewSelector(ctx)

		// Handle add item signal.
		selector.AddReceive(addChannel, func(c workflow.ReceiveChannel, _ bool) {
			var req activities.AddItemRequest
			c.Receive(ctx, &req)

			logger.Info("Adding item to cart via activity", "customerID", customerID, "goodID", req.GoodID)

			addItemCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second, //nolint:mnd
				Summary:             "Add item to cart",
				RetryPolicy:         ao.RetryPolicy,
			})

			err := workflow.ExecuteActivity(addItemCtx, "AddItem", req).Get(ctx, nil)
			if err != nil {
				logger.Error("Failed to add item", "error", err)
			}

			resetSessionTimeout()
		})

		// Handle remove item signal.
		selector.AddReceive(removeChannel, func(c workflow.ReceiveChannel, _ bool) {
			var req activities.RemoveItemRequest
			c.Receive(ctx, &req)

			logger.Info("Removing item from cart via activity", "customerID", customerID, "goodID", req.GoodID)

			removeItemCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second, //nolint:mnd
				Summary:             "Remove item from cart",
				RetryPolicy:         ao.RetryPolicy,
			})

			err := workflow.ExecuteActivity(removeItemCtx, "RemoveItem", req).Get(ctx, nil)
			if err != nil {
				logger.Error("Failed to remove item", "error", err)
			}

			resetSessionTimeout()
		})

		// Handle reset signal.
		selector.AddReceive(resetChannel, func(c workflow.ReceiveChannel, _ bool) {
			c.Receive(ctx, nil)

			logger.Info("Resetting cart via activity", "customerID", customerID)

			resetCartCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
				Summary:             "Reset cart (clear all items)",
				RetryPolicy:         ao.RetryPolicy,
			})

			err := workflow.ExecuteActivity(resetCartCtx, "ResetCart", activities.ResetCartRequest{
				CustomerID: customerID,
			}).Get(ctx, nil)
			if err != nil {
				logger.Error("Failed to reset cart", "error", err)
			}

			resetSessionTimeout()
		})

		// Handle session timeout.
		selector.AddFuture(sessionTimeout, func(f workflow.Future) {
			if err := f.Get(ctx, nil); err != nil {
				return
			}

			logger.Info("Cart session timed out, resetting cart", "customerID", customerID)

			timeoutResetCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
				Summary:             "Reset cart after session timeout",
				RetryPolicy:         ao.RetryPolicy,
			})
			//nolint:errcheck // best-effort reset on timeout, ignore activity result
			_ = workflow.ExecuteActivity(timeoutResetCtx, "ResetCart", activities.ResetCartRequest{
				CustomerID: customerID,
			}).Get(ctx, nil)

			resetSessionTimeout()
		})

		selector.Select(ctx)
	}
}
