package on_order_completed

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	logger "github.com/shortlink-org/go-sdk/logger"

	orderevents "github.com/shortlink-org/shop/oms/internal/domain/order/v1/events/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

type Handler struct {
	log             logger.Logger
	uow             ports.UnitOfWork
	orderRepo       ports.OrderRepository
	leaderboardRepo ports.LeaderboardRepository
}

func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
	leaderboardRepo ports.LeaderboardRepository,
) (*Handler, error) {
	return &Handler{
		log:             log,
		uow:             uow,
		orderRepo:       orderRepo,
		leaderboardRepo: leaderboardRepo,
	}, nil
}

func (h *Handler) Handle(ctx context.Context, event *orderevents.OrderCompleted) error {
	orderID, err := uuid.Parse(event.GetOrderId())
	if err != nil {
		return fmt.Errorf("parse order id: %w", err)
	}

	txCtx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := h.uow.Rollback(txCtx); rollbackErr != nil {
			h.log.Warn("leaderboard order-completed rollback failed", slog.String("error", rollbackErr.Error()))
		}
	}()

	orderState, err := h.orderRepo.Load(txCtx, orderID)
	if err != nil {
		return fmt.Errorf("load completed order: %w", err)
	}

	if err := h.uow.Commit(txCtx); err != nil {
		return fmt.Errorf("commit read transaction: %w", err)
	}
	committed = true

	items := orderState.GetItems()
	snapshotItems := make([]ports.LeaderboardOrderItem, 0, len(items))
	for _, item := range items {
		snapshotItems = append(snapshotItems, ports.LeaderboardOrderItem{
			GoodID:   item.GetGoodId(),
			Quantity: item.GetQuantity(),
			Price:    item.GetPrice(),
		})
	}

	applied, err := h.leaderboardRepo.ApplyOrderCompleted(ctx, ports.CompletedOrderLeaderboardSnapshot{
		OrderID:          orderID,
		AggregateVersion: event.GetAggregateVersion(),
		CompletedAt:      event.GetCompletedAt().AsTime(),
		Items:            snapshotItems,
	})
	if err != nil {
		return fmt.Errorf("apply leaderboard projection: %w", err)
	}

	if !applied {
		h.log.Info("leaderboard projection ignored duplicate completed order",
			slog.String("order_id", event.GetOrderId()),
			slog.Int64("aggregate_version", int64(event.GetAggregateVersion())))
	}

	return nil
}
