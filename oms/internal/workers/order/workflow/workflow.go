package order_workflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	v2 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/workers/order/activities"
)

// WorkflowInput contains all inputs for the order workflow.
type WorkflowInput struct {
	OrderID      uuid.UUID
	CustomerID   uuid.UUID
	Items        v2.Items
	DeliveryInfo *DeliveryInfoInput // Optional: if nil, no delivery is requested
}

// DeliveryInfoInput contains delivery info for the workflow.
type DeliveryInfoInput struct {
	PickupAddress   activities.DeliveryAddressDTO
	DeliveryAddress activities.DeliveryAddressDTO
	StartTime       time.Time
	EndTime         time.Time
	WeightKg        float64
	Dimensions      string
	Priority        int32
}

// Workflow is a Temporal workflow that orchestrates order processing.
// This workflow implements the saga pattern for order creation:
// 1. Create order in database
// 2. Reserve stock (TODO: implement stock service integration)
// 3. Process payment (TODO: implement payment service integration)
// 4. Request delivery (if delivery info provided)
// 5. Complete order
//
// On failure, compensation activities are executed to rollback changes.
// The workflow is deterministic - all side effects go through activities.
func Workflow(ctx workflow.Context, orderID, customerID uuid.UUID, items v2.Items) error {
	return WorkflowWithDelivery(ctx, WorkflowInput{
		OrderID:      orderID,
		CustomerID:   customerID,
		Items:        items,
		DeliveryInfo: nil, // No delivery info in legacy signature
	})
}

// WorkflowWithDelivery is the extended workflow that supports delivery integration.
func WorkflowWithDelivery(ctx workflow.Context, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)

	// Set initial workflow details for UI visibility
	workflow.SetCurrentDetails(ctx, fmt.Sprintf("Initializing order processing for %d items", len(input.Items)))

	// Activity options with retry policy
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				v2.ErrOrderItemsEmpty.Error(),
				v2.ErrOrderItemInvalid.Error(),
				v2.ErrOrderItemQuantityZero.Error(),
				v2.ErrOrderItemPriceNegative.Error(),
				v2.ErrOrderItemPriceZero.Error(),
				v2.ErrOrderItemsDuplicate.Error(),
				v2.ErrOrderInvalidStateTransition.Error(),
				v2.ErrInvalidOrderID.Error(),
				v2.ErrInvalidGoodID.Error(),
				v2.ErrInvalidOrderStatus.Error(),
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Track order status for queries
	var orderStatus string = "PROCESSING"
	var orderError error

	// Set up query handler for getting order status
	err := workflow.SetQueryHandler(ctx, v2.WorkflowQueryGet, func() (string, error) {
		return orderStatus, orderError
	})
	if err != nil {
		return err
	}

	// Signal channels
	cancelChannel := workflow.GetSignalChannel(ctx, v2.WorkflowSignalCancel)
	completeChannel := workflow.GetSignalChannel(ctx, v2.WorkflowSignalComplete)

	// Create a cancellable context for the saga
	sagaCtx, cancelSaga := workflow.WithCancel(ctx)

	// Run saga in a goroutine so we can handle signals
	sagaDone := workflow.NewChannel(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		sagaErr := executeSagaWithDelivery(sagaCtx, input)
		if sagaErr != nil {
			orderStatus = "FAILED"
			orderError = sagaErr
		} else {
			orderStatus = "COMPLETED"
		}
		sagaDone.Send(ctx, sagaErr)
	})

	// Wait for saga completion or signals
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, nil)
		logger.Info("Order cancellation signal received")
		cancelSaga()
		orderStatus = "CANCELLED"
	})

	selector.AddReceive(completeChannel, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, nil)
		logger.Info("Order completion signal received")
	})

	selector.AddReceive(sagaDone, func(c workflow.ReceiveChannel, _ bool) {
		var err error
		c.Receive(ctx, &err)
		if err != nil {
			logger.Error("Saga failed", "error", err)
		} else {
			logger.Info("Saga completed successfully")
		}
	})

	// Wait for first event
	selector.Select(ctx)

	return orderError
}

// executeSaga executes the order processing saga (legacy version without delivery).
// Returns error if any step fails (compensation should be handled).
func executeSaga(ctx workflow.Context, orderID, customerID uuid.UUID, items v2.Items) error {
	return executeSagaWithDelivery(ctx, WorkflowInput{
		OrderID:      orderID,
		CustomerID:   customerID,
		Items:        items,
		DeliveryInfo: nil,
	})
}

// executeSagaWithDelivery executes the order processing saga with optional delivery.
// Returns error if any step fails (compensation should be handled).
func executeSagaWithDelivery(ctx workflow.Context, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	hasDelivery := input.DeliveryInfo != nil
	totalSteps := 4
	if hasDelivery {
		totalSteps = 5
	}

	// Step 1: Create order in database (already done by usecase before workflow starts)
	// The order is created by the usecase that starts this workflow
	workflow.SetCurrentDetails(ctx, fmt.Sprintf("**Step 1/%d:** Order created in database ✓", totalSteps))
	logger.Info("Order already created in database", "orderID", input.OrderID)

	// Step 2: Reserve stock (TODO: implement stock service activity)
	workflow.SetCurrentDetails(ctx, fmt.Sprintf("**Step 2/%d:** Reserving stock...", totalSteps))
	// err := workflow.ExecuteActivity(ctx, activities.ReserveStock, activities.ReserveStockRequest{
	//     OrderID: input.OrderID,
	//     Items:   input.Items,
	// }).Get(ctx, nil)
	// if err != nil {
	//     workflow.SetCurrentDetails(ctx, "**Failed:** Stock reservation failed, compensating...")
	//     // Compensation: cancel order
	//     _ = workflow.ExecuteActivity(ctx, activities.CancelOrder, activities.CancelOrderRequest{OrderID: input.OrderID}).Get(ctx, nil)
	//     return err
	// }

	// Step 3: Process payment (TODO: implement payment service activity)
	workflow.SetCurrentDetails(ctx, fmt.Sprintf("**Step 3/%d:** Processing payment...", totalSteps))
	// err = workflow.ExecuteActivity(ctx, activities.ProcessPayment, ...).Get(ctx, nil)
	// if err != nil {
	//     workflow.SetCurrentDetails(ctx, "**Failed:** Payment processing failed, compensating...")
	//     // Compensation: release stock, cancel order
	//     _ = workflow.ExecuteActivity(ctx, activities.ReleaseStock, ...).Get(ctx, nil)
	//     _ = workflow.ExecuteActivity(ctx, activities.CancelOrder, ...).Get(ctx, nil)
	//     return err
	// }

	// Step 4: Request delivery (if delivery info provided)
	if hasDelivery {
		workflow.SetCurrentDetails(ctx, fmt.Sprintf("**Step 4/%d:** Requesting delivery...", totalSteps))

		var deliveryResp activities.RequestDeliveryResponse
		err := workflow.ExecuteActivity(ctx, "RequestDelivery", activities.RequestDeliveryRequest{
			OrderID:         input.OrderID,
			CustomerID:      input.CustomerID,
			PickupAddress:   input.DeliveryInfo.PickupAddress,
			DeliveryAddress: input.DeliveryInfo.DeliveryAddress,
			StartTime:       input.DeliveryInfo.StartTime,
			EndTime:         input.DeliveryInfo.EndTime,
			WeightKg:        input.DeliveryInfo.WeightKg,
			Dimensions:      input.DeliveryInfo.Dimensions,
			Priority:        input.DeliveryInfo.Priority,
		}).Get(ctx, &deliveryResp)
		if err != nil {
			workflow.SetCurrentDetails(ctx, "**Failed:** Delivery request failed, compensating...")
			logger.Error("Failed to request delivery", "error", err, "orderID", input.OrderID)
			// Compensation: cancel order (stock release would also be needed if implemented)
			var cancelActivities *activities.Activities
			_ = workflow.ExecuteActivity(ctx, cancelActivities.CancelOrder, activities.CancelOrderRequest{OrderID: input.OrderID}).Get(ctx, nil)
			return err
		}

		logger.Info("Delivery requested successfully",
			"orderID", input.OrderID,
			"packageID", deliveryResp.PackageID,
			"status", deliveryResp.Status)
	}

	// Final Step: Complete order
	currentStep := totalSteps
	workflow.SetCurrentDetails(ctx, fmt.Sprintf("**Step %d/%d:** Completing order...", currentStep, totalSteps))
	// Note: For now, we just log. In a real implementation, you'd call an activity.
	logger.Info("Order processing completed", "orderID", input.OrderID)

	workflow.SetCurrentDetails(ctx, "**Completed:** Order processed successfully ✓")

	return nil
}
