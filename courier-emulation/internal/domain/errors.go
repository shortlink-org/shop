package domain

import "errors"

// Domain errors for courier-emulation. Use errors.Is/As when mapping to HTTP/gRPC.
var (
	ErrRouteTooShort            = errors.New("route must have at least 2 points")
	ErrCourierNotFound          = errors.New("courier not found")
	ErrRouteCompleted           = errors.New("route completed")
	ErrCourierHasActiveDelivery = errors.New("courier already has an active delivery")
	ErrDeliveryNotFound         = errors.New("delivery not found")
	ErrUnknownPhase             = errors.New("unknown phase")
)
