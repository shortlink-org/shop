package on_stock_changed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/oms/internal/domain"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/websocket"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// Handler handles StockChangedEvent by adjusting affected carts.
type Handler struct {
	log        logger.Logger
	uow        ports.UnitOfWork
	cartRepo   ports.CartRepository
	goodsIndex ports.CartGoodsIndex
	notifier   *websocket.Notifier
}

// NewHandler creates a new on_stock_changed event handler.
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
		notifier:   nil,
	}, nil
}

// SetNotifier sets the websocket notifier for UI notifications.
func (h *Handler) SetNotifier(notifier *websocket.Notifier) {
	h.notifier = notifier
}

// Handle reacts to StockChangedEvent by removing out-of-stock items from carts.
func (h *Handler) Handle(ctx context.Context, event Event) error {
	// If stock is not zero, no action needed
	if event.NewQuantity > 0 {
		return nil
	}

	h.log.Info("Stock depleted for good", slog.String("good_id", event.GoodID.String()))

	// Get all customers that have this good in their cart using the index
	customerIDs, err := h.goodsIndex.GetCustomersWithGood(ctx, event.GoodID)
	if err != nil {
		h.log.Warn("Failed to get customers with good from index",
			slog.String("good_id", event.GoodID.String()),
			slog.String("error", err.Error()))

		return err
	}

	if len(customerIDs) == 0 {
		h.log.Info("No carts found with the out-of-stock item", slog.String("good_id", event.GoodID.String()))
		return nil
	}

	// Process each cart
	for _, customerID := range customerIDs {
		err := h.processCart(ctx, customerID, event.GoodID)
		if err != nil {
			h.log.Warn("Failed to process cart",
				slog.String("customer_id", customerID.String()),
				slog.String("good_id", event.GoodID.String()),
				slog.String("error", err.Error()))
			// Continue processing other carts
		}
	}

	return nil
}

// processCart handles removing the out-of-stock item from a single cart.
func (h *Handler) processCart(ctx context.Context, customerID, goodID uuid.UUID) error {
	// Begin transaction
	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err := h.uow.Rollback(ctx)
		if err != nil {
			h.log.Warn("transaction rollback failed", slog.Any("error", err))
		}
	}()

	// Load cart
	cart, err := h.cartRepo.Load(ctx, customerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Cart doesn't exist, clean up index
			removeErr := h.goodsIndex.RemoveGoodFromCart(ctx, goodID, customerID)
			if removeErr != nil {
				h.log.Warn("failed to remove good from cart index", slog.String("error", removeErr.Error()))
			}

			return nil
		}

		return err
	}

	// Find the item and its quantity
	var itemQuantity int32

	found := false

	for _, item := range cart.GetItems() {
		if item.GetGoodId() == goodID {
			itemQuantity = item.GetQuantity()
			found = true

			break
		}
	}

	if !found {
		// Item was already removed, clean up index
		removeErr := h.goodsIndex.RemoveGoodFromCart(ctx, goodID, customerID)
		if removeErr != nil {
			h.log.Warn("failed to remove good from cart index", slog.String("error", removeErr.Error()))
		}

		return nil
	}

	h.log.Info("Removing out-of-stock item from cart",
		slog.String("customer_id", customerID.String()),
		slog.String("good_id", goodID.String()),
		slog.Int("quantity", int(itemQuantity)))

	// Create cart item for removal
	cartItem, err := itemv1.NewItem(goodID, itemQuantity)
	if err != nil {
		return fmt.Errorf("failed to construct cart item: %w", err)
	}

	// Remove the item from cart
	if err := cart.RemoveItem(cartItem); err != nil {
		return err
	}

	// Save cart
	if err := h.cartRepo.Save(ctx, cart); err != nil {
		return err
	}

	// Commit transaction
	if err := h.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Post-commit: use a context without tx so index/notifier don't see a closed transaction
	ctxClean, cancel := uow.ContextWithoutTx(ctx)
	defer cancel()

	// Update index
	if err := h.goodsIndex.RemoveGoodFromCart(ctxClean, goodID, customerID); err != nil {
		h.log.Warn("failed to update cart goods index", slog.Any("error", err))
	}

	// Send websocket notification to UI
	if h.notifier != nil {
		err := h.notifier.NotifyStockDepleted(customerID, goodID)
		if err != nil {
			h.log.Warn("Failed to send websocket notification",
				slog.String("customer_id", customerID.String()),
				slog.String("good_id", goodID.String()),
				slog.String("error", err.Error()))
		}
	}

	return nil
}
