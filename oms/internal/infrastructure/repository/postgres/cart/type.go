package postgres

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/schema/crud"
)

const (
	// Cache configuration for L1 in-memory cache
	cacheNumCounters = 10_000      // track 10k carts
	cacheMaxCost     = 100_000_000 // ~100MB
	cacheBufferItems = 64
	cacheTTL         = 10 * time.Second // short TTL for eventual consistency
)

// Store implements CartRepository using PostgreSQL with L1 Ristretto cache.
type Store struct {
	client *pgxpool.Pool
	query  *crud.Queries
	cache  *ristretto.Cache[string, *cart.State]
}
