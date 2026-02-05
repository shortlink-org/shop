package domain

import "errors"

// Domain-level errors for pricer. Use errors.Is/As when mapping to gRPC/HTTP.
var (
	// ErrInvalidCart is returned when cart data is invalid (e.g. malformed IDs or prices).
	ErrInvalidCart = errors.New("invalid cart")
)
