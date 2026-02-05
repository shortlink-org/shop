package v1

import (
	"errors"
	"fmt"

	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common"
)

// Sentinel domain errors for order aggregate. Handlers can use errors.Is/As to map to gRPC/HTTP codes.
var (
	ErrInvalidDeliveryInfo = errors.New("invalid delivery info: address, delivery period and package info are required")
)

// ErrOrderTerminalState is returned when an operation is not allowed because the order is in a terminal state (COMPLETED or CANCELLED).
type ErrOrderTerminalState struct {
	Status OrderStatus
}

func (e *ErrOrderTerminalState) Error() string {
	return fmt.Sprintf("order in terminal state: %s", e.Status)
}

// ErrDeliveryAlreadyInProgress is returned when delivery info cannot be updated because the package is already assigned or in transit.
type ErrDeliveryAlreadyInProgress struct {
	DeliveryStatus commonv1.DeliveryStatus
}

func (e *ErrDeliveryAlreadyInProgress) Error() string {
	return fmt.Sprintf("cannot update delivery info: package already %s", e.DeliveryStatus)
}

// ErrInvalidDeliveryStatusTransition is returned when the delivery status transition is not allowed (e.g. UNSPECIFIED -> DELIVERED).
type ErrInvalidDeliveryStatusTransition struct {
	From commonv1.DeliveryStatus
	To   commonv1.DeliveryStatus
}

func (e *ErrInvalidDeliveryStatusTransition) Error() string {
	return fmt.Sprintf("invalid delivery status transition from %s to %s", e.From, e.To)
}

// ErrInvalidOrderTransition is returned when an order state transition is not allowed (e.g. CompleteOrder requires PROCESSING).
type ErrInvalidOrderTransition struct {
	From OrderStatus
	To   OrderStatus
}

func (e *ErrInvalidOrderTransition) Error() string {
	return fmt.Sprintf("cannot transition from %s to %s", e.From, e.To)
}
