package kafka

import "errors"

// Validation errors for NewDeliverOrderEvent. Callers can use errors.Is.
var (
	ErrReasonMustBeEmpty = errors.New("reason must be empty when delivered")
	ErrReasonRequired    = errors.New("reason is required when not delivered")
	ErrInvalidReason     = errors.New("invalid not_delivered reason")
)
