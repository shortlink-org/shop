package ports

import "errors"

var (
	// ErrNotFound is returned when an aggregate is not found in the repository.
	ErrNotFound = errors.New("aggregate not found")

	// ErrVersionConflict is returned when optimistic locking detects a version mismatch.
	ErrVersionConflict = errors.New("optimistic lock: version conflict")
)
