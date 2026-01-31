//go:generate sqlc generate -f ./schema/sqlc.yaml

package postgres

import (
	"context"
	"embed"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shortlink-org/go-sdk/db"
	"github.com/shortlink-org/go-sdk/db/drivers/postgres/migrate"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart/schema/crud"
)

var (
	//go:embed migrations/*.sql
	migrations embed.FS
)

// New creates a new PostgreSQL cart repository.
func New(ctx context.Context, store db.DB) (*Store, error) {
	client, ok := store.GetConn().(*pgxpool.Pool)
	if !ok {
		return nil, db.ErrGetConnection
	}

	// Run migrations
	if err := migrate.Migration(ctx, store, migrations, "repository_cart"); err != nil {
		return nil, err
	}

	return &Store{
		client: client,
		query:  crud.New(client),
	}, nil
}
