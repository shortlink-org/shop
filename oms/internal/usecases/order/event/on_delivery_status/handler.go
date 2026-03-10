package on_delivery_status

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/logger"

	deliverycommon "github.com/shortlink-org/shop/oms/internal/domain/delivery/common/v1"
	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/kafka"
)

// Handler handles delivery status events.
type Handler struct {
	log       logger.Logger
	uow       ports.UnitOfWork
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

// NewHandler creates a new delivery status event handler.
func NewHandler(
	log logger.Logger,
	uow ports.UnitOfWork,
	orderRepo ports.OrderRepository,
	publisher ports.EventPublisher,
) (*Handler, error) {
	return &Handler{
		log:       log,
		uow:       uow,
		orderRepo: orderRepo,
		publisher: publisher,
	}, nil
}

// HandleDeliveryStatus processes a delivery status event.
// Pattern: Begin -> Load -> Mutate -> Save -> Publish in tx -> Commit.
func (h *Handler) HandleDeliveryStatus(ctx context.Context, event kafka.DeliveryStatusEvent) error {
	h.log.Info("Processing delivery status event",
		slog.String("package_id", event.PackageID),
		slog.String("order_id", event.OrderID),
		slog.String("status", event.Status),
		slog.String("event_type", string(event.EventType)))

	ctx, err := h.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}

		rollbackErr := h.uow.Rollback(ctx)
		if rollbackErr != nil {
			h.log.Warn("transaction rollback failed", slog.Any("error", rollbackErr))
		}
	}()

	order, err := h.loadOrderForEvent(ctx, event)
	if err != nil {
		return err
	}

	if isDuplicateOrStale(order, event.EventType) {
		h.log.Info("Ignoring duplicate or stale delivery event",
			slog.String("package_id", event.PackageID),
			slog.String("order_id", order.GetOrderID().String()),
			slog.String("event_type", string(event.EventType)),
			slog.String("current_delivery_status", order.GetDeliveryStatus().String()))

		if err := h.uow.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit noop transaction: %w", err)
		}
		committed = true

		return nil
	}

	if err := applyDeliveryEvent(order, event); err != nil {
		return fmt.Errorf("failed to apply delivery event: %w", err)
	}

	if err := h.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	for _, domainEvent := range order.GetDomainEvents() {
		if err := h.publisher.Publish(ctx, domainEvent); err != nil {
			return fmt.Errorf("failed to publish domain event to outbox: %w", err)
		}
	}

	if err := h.uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	order.ClearDomainEvents()

	h.log.Info("Successfully processed delivery status event",
		slog.String("order_id", order.GetOrderID().String()),
		slog.String("new_delivery_status", order.GetDeliveryStatus().String()))

	return nil
}

func (h *Handler) loadOrderForEvent(ctx context.Context, event kafka.DeliveryStatusEvent) (*orderv1.OrderState, error) {
	if event.OrderID != "" {
		orderID, err := uuid.Parse(event.OrderID)
		if err != nil {
			return nil, fmt.Errorf("invalid order_id: %w", err)
		}

		order, err := h.orderRepo.Load(ctx, orderID)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				return nil, fmt.Errorf("order not found for delivery event: %w", err)
			}

			return nil, fmt.Errorf("failed to load order: %w", err)
		}

		return order, nil
	}

	packageID, err := parseRequiredUUID("package_id", event.PackageID)
	if err != nil {
		return nil, err
	}

	order, err := h.orderRepo.LoadByPackageID(ctx, *packageID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("order not found for package_id %s: %w", event.PackageID, err)
		}

		return nil, fmt.Errorf("failed to load order by package_id: %w", err)
	}

	return order, nil
}

func applyDeliveryEvent(order *orderv1.OrderState, event kafka.DeliveryStatusEvent) error {
	packageID, err := parseOptionalUUID("package_id", event.PackageID)
	if err != nil {
		return err
	}

	courierID, err := parseOptionalUUID("courier_id", event.CourierID)
	if err != nil {
		return err
	}

	switch event.EventType {
	case kafka.EventTypePackageAccepted:
		return order.ApplyDeliveryAccepted(packageID, event.OccurredAt)
	case kafka.EventTypePackageAssigned:
		return order.ApplyDeliveryAssigned(packageID, courierID, event.OccurredAt)
	case kafka.EventTypePackageInTransit:
		return order.ApplyDeliveryInTransit(packageID, courierID, event.OccurredAt)
	case kafka.EventTypePackageDelivered:
		return order.ApplyDeliveryDelivered(
			packageID,
			courierID,
			mapDeliveryLocation(event.DeliveryLocation),
			event.OccurredAt,
		)
	case kafka.EventTypePackageNotDelivered:
		return order.ApplyDeliveryFailed(
			packageID,
			courierID,
			mapNotDeliveredDetails(event.NotDeliveredDetails),
			event.OccurredAt,
		)
	default:
		return fmt.Errorf("unsupported delivery event type: %s", event.EventType)
	}
}

func isDuplicateOrStale(order *orderv1.OrderState, eventType kafka.DeliveryEventType) bool {
	currentStatus := order.GetDeliveryStatus()
	targetStatus, ok := targetDeliveryStatus(eventType)
	if !ok {
		return false
	}

	if currentStatus == targetStatus {
		return true
	}

	if currentStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED &&
		targetStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED {
		return false
	}

	if currentStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED &&
		targetStatus == commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED {
		return false
	}

	return deliveryStatusRank(currentStatus) > deliveryStatusRank(targetStatus)
}

func targetDeliveryStatus(eventType kafka.DeliveryEventType) (commonv1.DeliveryStatus, bool) {
	switch eventType {
	case kafka.EventTypePackageAccepted:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED, true
	case kafka.EventTypePackageAssigned:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED, true
	case kafka.EventTypePackageInTransit:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT, true
	case kafka.EventTypePackageDelivered:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED, true
	case kafka.EventTypePackageNotDelivered:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED, true
	default:
		return commonv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED, false
	}
}

func deliveryStatusRank(status commonv1.DeliveryStatus) int {
	switch status {
	case commonv1.DeliveryStatus_DELIVERY_STATUS_ACCEPTED:
		return 1
	case commonv1.DeliveryStatus_DELIVERY_STATUS_ASSIGNED:
		return 2
	case commonv1.DeliveryStatus_DELIVERY_STATUS_IN_TRANSIT:
		return 3
	case commonv1.DeliveryStatus_DELIVERY_STATUS_DELIVERED,
		commonv1.DeliveryStatus_DELIVERY_STATUS_NOT_DELIVERED:
		return 4
	default:
		return 0
	}
}

func parseRequiredUUID(fieldName, value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, fmt.Errorf("%s is required", fieldName)
	}

	return parseOptionalUUID(fieldName, value)
}

func parseOptionalUUID(fieldName, value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", fieldName, err)
	}

	return &parsed, nil
}

func mapDeliveryLocation(location *deliverycommon.Location) *commonv1.DeliveryLocation {
	if location == nil {
		return nil
	}

	return &commonv1.DeliveryLocation{
		Latitude:  location.GetLatitude(),
		Longitude: location.GetLongitude(),
		Accuracy:  location.GetAccuracy(),
		Timestamp: location.GetTimestamp(),
	}
}

func mapNotDeliveredDetails(details *deliverycommon.NotDeliveredDetails) *commonv1.NotDeliveredDetails {
	if details == nil {
		return nil
	}

	return &commonv1.NotDeliveredDetails{
		Reason:      commonv1.NotDeliveredReason(details.GetReason()),
		Description: details.GetDescription(),
	}
}
