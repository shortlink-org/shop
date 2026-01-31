package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	v2 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	v1 "github.com/shortlink-org/shop/oms/internal/domain/queue/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/temporal"
)

// OrderWorkflowName is the registered name of the order processing workflow.
// This constant is used to start the workflow without importing the workflow package,
// which would create an import cycle.
const OrderWorkflowName = "OrderWorkflow"

// Create creates a new order and persists it to the database.
// For complex order processing (sagas), a Temporal workflow is started.
// deliveryInfo is optional (nil = self-pickup).
func (uc *UC) Create(ctx context.Context, orderID uuid.UUID, customerID uuid.UUID, items v2.Items, deliveryInfo *v2.DeliveryInfo) error {
	// Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = uc.uow.Rollback(ctx) }()

	// 1. Create domain aggregate
	order := v2.NewOrderState(customerID)
	order.SetID(orderID)

	// 2. Apply business logic (create order with items)
	if err := order.CreateOrder(items); err != nil {
		return err
	}

	// 3. Set delivery info if provided
	if deliveryInfo != nil {
		order.SetDeliveryInfo(*deliveryInfo)
	}

	// 4. Persist to database
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	// Commit transaction before starting workflow
	if err := uc.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 5. Start Temporal workflow for order processing (saga pattern)
	// Note: We use workflow name string to avoid import cycle.
	// The workflow is registered in the worker with this name.
	workflowID := fmt.Sprintf("order-%s", orderID.String())
	_, err = uc.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: temporal.GetQueueName(v1.OrderTaskQueue),
	}, OrderWorkflowName, orderID, customerID, items, deliveryInfo)
	if err != nil {
		// Order is persisted, but workflow failed to start
		// TODO: Consider compensation or retry logic
		uc.log.Error("failed to start order workflow",
			slog.String("orderID", orderID.String()),
			slog.Any("error", err))
		return err
	}

	return nil
}

// CreateInDB creates an order directly in the database without starting a workflow.
// This is used by Temporal activities when the workflow orchestration calls this.
// deliveryInfo is optional (nil = self-pickup).
func (uc *UC) CreateInDB(ctx context.Context, orderID uuid.UUID, customerID uuid.UUID, items v2.Items, deliveryInfo *v2.DeliveryInfo) error {
	// Begin transaction
	ctx, err := uc.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = uc.uow.Rollback(ctx) }()

	// 1. Create domain aggregate
	order := v2.NewOrderState(customerID)
	order.SetID(orderID)

	// 2. Apply business logic
	if err := order.CreateOrder(items); err != nil {
		return err
	}

	// 3. Set delivery info if provided
	if deliveryInfo != nil {
		order.SetDeliveryInfo(*deliveryInfo)
	}

	// 4. Persist to database
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	// Commit transaction
	return uc.uow.Commit(ctx)
}
