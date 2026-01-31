package order_worker

import (
	"context"

	logger "github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	"github.com/shortlink-org/shop/oms/internal/usecases/order"
	order_workflow "github.com/shortlink-org/shop/oms/internal/workers/order/workflow"
)

func New(ctx context.Context, c client.Client, log logger.Logger) (worker.Worker, error) {
	// This worker hosts Workflow functions
	// Activities are registered separately when order UC is available
	w := worker.New(c, temporal.GetQueueName(v1.OrderTaskQueue), worker.Options{})

	// Register workflow with a specific name to avoid import cycles
	// The name must match order.OrderWorkflowName used in usecases/order/create.go
	w.RegisterWorkflowWithOptions(order_workflow.Workflow, workflow.RegisterOptions{
		Name: order.OrderWorkflowName,
	})

	// Note: Activities should be registered by the main application
	// when the order UC is available, using RegisterActivities()

	// Start listening to the Task Queue
	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			panic(err)
		}
	}()

	log.Info("Worker started")

	return w, nil
}
