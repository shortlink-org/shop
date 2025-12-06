package v1

import (
	commonv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1/common/v1"
)

// Re-export generated proto enums/structs that are part of the ubiquitous
// language so the domain layer can stay self-contained while still sharing
// definitions with the transport layer.
type (
	// OrderStatus describes the lifecycle of an order.
	OrderStatus = commonv1.OrderStatus
	// DeliveryAddress represents an address with coordinates.
	DeliveryAddress = commonv1.DeliveryAddress
)

const (
	OrderStatus_ORDER_STATUS_UNSPECIFIED OrderStatus = commonv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	OrderStatus_ORDER_STATUS_PENDING     OrderStatus = commonv1.OrderStatus_ORDER_STATUS_PENDING
	OrderStatus_ORDER_STATUS_PROCESSING  OrderStatus = commonv1.OrderStatus_ORDER_STATUS_PROCESSING
	OrderStatus_ORDER_STATUS_COMPLETED   OrderStatus = commonv1.OrderStatus_ORDER_STATUS_COMPLETED
	OrderStatus_ORDER_STATUS_CANCELLED   OrderStatus = commonv1.OrderStatus_ORDER_STATUS_CANCELLED
)

var (
	// OrderStatus_name allows mapping FSM state strings back to domain enums.
	OrderStatus_name = commonv1.OrderStatus_name
	// OrderStatus_value is preserved for completeness when interacting with
	// generated code.
	OrderStatus_value = commonv1.OrderStatus_value
)
