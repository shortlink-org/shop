package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/crud"
)

// Store implements OrderRepository using PostgreSQL.
type Store struct {
	client *pgxpool.Pool
	query  *crud.Queries
}
