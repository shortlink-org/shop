package on_delivery_status

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/logger"

	common "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
)

// Handler handles delivery status events.
type Handler struct {
	log       logger.Logger
	orderRepo ports.OrderRepository
}

// NewHandler creates a new delivery status event handler.
func NewHandler(
	log logger.Logger,
	orderRepo ports.OrderRepository,
) (*Handler, error) {
	return &Handler{
		log:       log,
		orderRepo: orderRepo,
	}, nil
}

// HandleDeliveryStatus processes a delivery status event.
// Updates the order's delivery status based on the event.
func (h *Handler) HandleDeliveryStatus(ctx context.Context, event kafka.DeliveryStatusEvent) error {
	h.log.Info("Processing delivery status event",
		slog.String("package_id", event.PackageID),
		slog.String("order_id", event.OrderID),
		slog.String("status", event.Status),
		slog.String("event_type", string(event.EventType)))

	// Parse order ID
	orderID, err := uuid.Parse(event.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order_id: %w", err)
	}

	// Load order
	order, err := h.orderRepo.Load(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to load order: %w", err)
	}

	if order == nil {
		h.log.Warn("Order not found for delivery event",
			slog.String("order_id", event.OrderID))

		return nil // Don't retry for non-existent orders
	}

	// Parse package ID if needed to update it
	if event.PackageID != "" {
		packageID, parseErr := uuid.Parse(event.PackageID)
		if parseErr == nil {
			// Update package ID if not already set
			deliveryInfo := order.GetDeliveryInfo()
			if deliveryInfo != nil && deliveryInfo.GetPackageId() == nil {
				deliveryInfo.SetPackageId(packageID)
			}
		}
	}

	// Update delivery status based on event type
	switch event.EventType {
	case kafka.EventTypePackageInTransit:
		err := order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT)
		if err != nil {
			h.log.Warn("Failed to set delivery status to IN_TRANSIT",
				slog.Any("error", err),
				slog.String("order_id", event.OrderID))
		}

	case kafka.EventTypePackageDelivered:
		err := order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_DELIVERED)
		if err != nil {
			h.log.Warn("Failed to set delivery status to DELIVERED",
				slog.Any("error", err),
				slog.String("order_id", event.OrderID))
		}
		// Mark order as completed
		err = order.CompleteOrder()
		if err != nil {
			h.log.Warn("Failed to complete order",
				slog.Any("error", err),
				slog.String("order_id", event.OrderID))
		}

	case kafka.EventTypePackageNotDelivered:
		err := order.SetDeliveryStatus(common.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED)
		if err != nil {
			h.log.Warn("Failed to set delivery status to NOT_DELIVERED",
				slog.Any("error", err),
				slog.String("order_id", event.OrderID),
				slog.String("reason", event.Reason))
		}

	default:
		h.log.Debug("Ignoring unknown delivery event type",
			slog.String("event_type", string(event.EventType)))

		return nil
	}

	// Save order
	if err := h.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	h.log.Info("Successfully processed delivery status event",
		slog.String("order_id", event.OrderID),
		slog.String("new_status", event.Status))

	return nil
}
