package order_worker

import (
	"context"

	logger "github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	temporalInfra "github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	order_workflow "github.com/shortlink-org/shop/oms/internal/workers/order/workflow"
)

// OrderWorker is a wrapper type for Order Temporal worker (needed for wire DI disambiguation)
type OrderWorker struct {
	Worker worker.Worker
}

func New(ctx context.Context, c client.Client, log logger.Logger) (OrderWorker, error) {
	// This worker hosts Workflow functions
	// Activities are registered separately when order UC is available
	w := worker.New(c, temporalInfra.GetQueueName(v1.OrderTaskQueue), worker.Options{})

	// Register workflow with a specific name to avoid import cycles
	// The name must match temporalEvents.OrderWorkflowName used in infrastructure/events/temporal
	w.RegisterWorkflowWithOptions(order_workflow.Workflow, workflow.RegisterOptions{
		Name: temporalInfra.OrderWorkflowName,
	})

	// Note: Activities should be registered by the main application
	// when the order handlers are available, using RegisterActivities()

	// Start listening to the Task Queue
	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			panic(err)
		}
	}()

	log.Info("Worker started")

	return OrderWorker{Worker: w}, nil
}
