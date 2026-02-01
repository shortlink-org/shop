package postgres

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
)

const (
	// Cache configuration for L1 in-memory cache
	cacheNumCounters = 50_000       // track 50k orders
	cacheMaxCost     = 200_000_000  // ~200MB
	cacheBufferItems = 64
	cacheTTL         = 5 * time.Minute // orders are mostly immutable
)

// Store implements OrderRepository using PostgreSQL with L1 Ristretto cache.
type Store struct {
	client *pgxpool.Pool
	query  *crud.Queries
	cache  *ristretto.Cache[string, *order.OrderState]
}
