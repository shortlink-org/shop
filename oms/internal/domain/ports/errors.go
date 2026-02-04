package ports

import "github.com/shortlink-org/shop/oms/internal/domain"

// Re-export domain errors so repository interfaces can document "returns domain.ErrNotFound"
// and callers can use errors.Is(err, ports.ErrNotFound). The canonical definitions are in domain.
var (
	ErrNotFound        = domain.ErrNotFound
	ErrVersionConflict = domain.ErrVersionConflict
	ErrConflict        = domain.ErrConflict
	ErrValidation      = domain.ErrValidation
	ErrUnavailable     = domain.ErrUnavailable
)
