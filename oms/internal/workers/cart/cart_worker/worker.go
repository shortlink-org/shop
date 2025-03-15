package cart_worker

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/shortlink-org/shortlink/boundaries/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shortlink/boundaries/shop/oms/internal/workers/cart/workflow"
	"github.com/shortlink-org/shortlink/pkg/logger"
)

func New(ctx context.Context, c client.Client, log logger.Logger) (worker.Worker, error) {
	// This worker hosts both Worker and Activity functions
	w := worker.New(c, v1.CART_TASK_QUEUE, worker.Options{})

	w.RegisterWorkflow(cart_workflow.Workflow)

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
