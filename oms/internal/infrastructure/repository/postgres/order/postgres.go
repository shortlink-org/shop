//go:generate sqlc generate -f ./schema/sqlc.yaml

package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/dgraph-io/ristretto/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shortlink-org/go-sdk/db"
	"github.com/shortlink-org/go-sdk/db/drivers/postgres/migrate"

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
)

var (
	//go:embed migrations/*.sql
	migrations embed.FS
)

// New creates a new PostgreSQL order repository with L1 cache.
func New(ctx context.Context, store db.DB) (*Store, error) {
	client, ok := store.GetConn().(*pgxpool.Pool)
	if !ok {
		return nil, db.ErrGetConnection
	}

	// Run migrations
	if err := migrate.Migration(ctx, store, migrations, "repository_order"); err != nil {
		return nil, err
	}

	// Initialize L1 cache
	cache, err := ristretto.NewCache(&ristretto.Config[string, *order.OrderState]{
		NumCounters: cacheNumCounters,
		MaxCost:     cacheMaxCost,
		BufferItems: cacheBufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create order cache: %w", err)
	}

	return &Store{
		client: client,
		query:  crud.New(client),
		cache:  cache,
	}, nil
}

// Close closes the repository and releases resources.
func (s *Store) Close() {
	if s.cache != nil {
		s.cache.Close()
	}
}
