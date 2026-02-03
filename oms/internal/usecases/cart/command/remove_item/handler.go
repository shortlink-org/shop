package remove_item

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Handler handles RemoveItem commands.
type Handler struct {
	log        logger.Logger
	uow        ports.UnitOfWork
	cartRepo   ports.CartRepository
	goodsIndex ports.CartGoodsIndex
}

// NewHandler creates a new RemoveItem handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	cartRepo ports.CartRepository,
	goodsIndex ports.CartGoodsIndex,
) (*Handler, error) {
	return &Handler{
		log:        log,
		uow:        uow,
		cartRepo:   cartRepo,
		goodsIndex: goodsIndex,
	}, nil
}

// Handle executes the RemoveItem command.
// Pattern: Load -> Domain method -> Save
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = h.uow.Rollback(ctx) }()

	// 1. Load aggregate
	cart, err := h.cartRepo.Load(ctx, cmd.CustomerID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Cart doesn't exist, nothing to remove
			return nil
		}
		return err
	}

	// 2. Call domain method (business logic)
	if err := cart.RemoveItem(cmd.Item); err != nil {
		return err
	}

	// 3. Save aggregate
	if err := h.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Commit transaction
	if err := h.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update index
	if err := h.goodsIndex.RemoveGoodFromCart(ctx, cmd.Item.GetGoodId(), cmd.CustomerID); err != nil {
		// Log but don't fail - index is eventually consistent
		h.log.Warn("failed to update cart goods index", slog.Any("error", err))
	}

	return nil
}

// Ensure Handler implements CommandHandler interface.
var _ ports.CommandHandler[Command] = (*Handler)(nil)
