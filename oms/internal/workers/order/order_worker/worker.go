package order_worker

import (
	"context"

	logger "github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	temporalInfra "github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	"github.com/shortlink-org/shop/oms/internal/workers/order/activities"
	order_workflow "github.com/shortlink-org/shop/oms/internal/workers/order/workflow"
)

// OrderWorker is a wrapper type for Order Temporal worker (needed for wire DI disambiguation)
type OrderWorker struct {
	Worker worker.Worker
}

// New creates a basic order worker without activities (for standalone worker service).
func New(ctx context.Context, c client.Client, log logger.Logger) (OrderWorker, error) {
	return NewWithActivities(ctx, c, log, nil)
}

// NewWithActivities creates an order worker with activities (for full OMS service).
func NewWithActivities(ctx context.Context, c client.Client, log logger.Logger, acts *activities.Activities) (OrderWorker, error) {
	// This worker hosts Workflow functions and Activities
	w := worker.New(c, temporalInfra.GetQueueName(v1.OrderTaskQueue), worker.Options{})

	// Register workflow with a specific name to avoid import cycles
	// The name must match temporalEvents.OrderWorkflowName used in infrastructure/events/temporal
	w.RegisterWorkflowWithOptions(order_workflow.Workflow, workflow.RegisterOptions{
		Name: temporalInfra.OrderWorkflowName,
	})

	// Register activities (only if provided)
	if acts != nil {
		w.RegisterActivity(acts.CancelOrder)
		w.RegisterActivity(acts.GetOrder)
		w.RegisterActivity(acts.RequestDelivery)
		log.Info("Order worker started with activities")
	} else {
		log.Info("Order worker started without activities (workflow-only mode)")
	}

	// Start listening to the Task Queue
	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			panic(err)
		}
	}()

	return OrderWorker{Worker: w}, nil
}
