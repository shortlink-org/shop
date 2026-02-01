package cart_worker

import (
	"context"

	logger "github.com/shortlink-org/go-sdk/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	cart_workflow "github.com/shortlink-org/shop/oms/internal/workers/cart/workflow"
)

// CartWorker is a wrapper type for Cart Temporal worker (needed for wire DI disambiguation)
type CartWorker struct {
	Worker worker.Worker
}

func New(ctx context.Context, c client.Client, log logger.Logger) (CartWorker, error) {
	// This worker hosts both Worker and Activity functions
	w := worker.New(c, temporal.GetQueueName(v1.CartTaskQueue), worker.Options{})

	w.RegisterWorkflow(cart_workflow.Workflow)

	// Start listening to the Task Queue
	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			panic(err)
		}
	}()

	log.Info("Worker started")

	return CartWorker{Worker: w}, nil
}
