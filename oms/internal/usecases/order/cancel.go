package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	domain "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
)

// Cancel cancels an order using the pattern: Load -> domain method -> Save
// Also signals the Temporal workflow if one exists.
func (uc *UC) Cancel(ctx context.Context, orderID uuid.UUID) error {
	// 1. Load aggregate
	order, err := uc.orderRepo.Load(ctx, orderID)
	if err != nil {
		return err
	}

	// 2. Call domain method (business logic)
	if err := order.CancelOrder(); err != nil {
		return err
	}

	// 3. Save aggregate
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	// 4. Signal Temporal workflow (if running)
	workflowID := fmt.Sprintf("order-%s", orderID.String())
	if err := uc.temporalClient.SignalWorkflow(ctx, workflowID, "", domain.WorkflowSignalCancel, nil); err != nil {
		// Log but don't fail - the order is already cancelled in DB
		uc.log.Warn("failed to signal cancel to workflow",
			slog.String("orderID", orderID.String()),
			slog.Any("error", err))
	}

	return nil
}

// CancelInDB cancels an order directly in the database without signaling workflow.
// This is used by Temporal activities when the workflow calls this for compensation.
func (uc *UC) CancelInDB(ctx context.Context, orderID uuid.UUID) error {
	// 1. Load aggregate
	order, err := uc.orderRepo.Load(ctx, orderID)
	if err != nil {
		return err
	}

	// 2. Call domain method
	if err := order.CancelOrder(); err != nil {
		return err
	}

	// 3. Save aggregate
	return uc.orderRepo.Save(ctx, order)
}
