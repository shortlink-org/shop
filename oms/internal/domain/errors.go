package domain

import (
	"errors"
	"fmt"
)

// Domain/application error sentinels. Used by usecases and API layer.
// Infrastructure errors (pgx, timeouts, network) should be wrapped into these
// with Unwrap for diagnostics.
var (
	// ErrNotFound is returned when an aggregate is not found in the repository.
	ErrNotFound = errors.New("aggregate not found")

	// ErrVersionConflict is returned when optimistic locking detects a version mismatch.
	ErrVersionConflict = errors.New("optimistic lock: version conflict")

	// ErrConflict is returned when the operation conflicts with current state.
	ErrConflict = errors.New("conflict")

	// ErrValidation is returned when input or domain validation fails.
	ErrValidation = errors.New("validation error")

	// ErrUnavailable is returned when an infrastructure failure occurs (db, network, timeout).
	// The underlying cause is available via errors.Unwrap for logging.
	ErrUnavailable = errors.New("unavailable")
)

// WrapUnavailable wraps an infrastructure error as ErrUnavailable, preserving the cause for Unwrap.
// Use in usecases when mapping infra failures (tx begin, commit, repo, network) to domain errors.
func WrapUnavailable(op string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w: %s: %w", ErrUnavailable, op, err)
}

// WrapValidation wraps a validation/domain error as ErrValidation, preserving the cause for Unwrap.
func WrapValidation(op string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w: %s: %w", ErrValidation, op, err)
}

// MapInfraErr returns err as-is if it is already a known domain error (ErrNotFound, ErrVersionConflict, etc.),
// otherwise wraps it as ErrUnavailable with op for usecase-layer mapping of infrastructure failures.
func MapInfraErr(op string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrVersionConflict) || errors.Is(err, ErrConflict) ||
		errors.Is(err, ErrValidation) || errors.Is(err, ErrUnavailable) {

		return err
	}

	return WrapUnavailable(op, err)
}
