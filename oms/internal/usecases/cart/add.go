package cart

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	v1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	v2 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
	"github.com/shortlink-org/shop/oms/internal/usecases/cart/dto"
	cart_workflow "github.com/shortlink-org/shop/oms/internal/workers/cart/workflow"
)

// Add adds an item to the cart.
func (uc *UC) Add(ctx context.Context, in *v1.State) error {
	workflowId := fmt.Sprintf("cart-%s", in.GetCustomerId().String())

	_, err := uc.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowId,
		TaskQueue: temporal.GetQueueName(v2.CartTaskQueue),
	}, cart_workflow.Workflow, in.GetCustomerId())
	if err != nil {
		return err
	}

	request := dto.CartStateToCartEvent(in, v1.Event_EVENT_ADD)
	err = uc.temporalClient.SignalWorkflow(ctx, workflowId, "", v1.Event_EVENT_ADD.String(), request)
	if err != nil {
		return err
	}

	// Update index: add goods to customer's cart index
	for _, item := range in.GetItems() {
		uc.goodsIndex.AddGoodToCart(item.GetGoodId(), in.GetCustomerId())
	}

	return nil
}
