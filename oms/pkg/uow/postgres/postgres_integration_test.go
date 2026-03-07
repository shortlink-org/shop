//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/testhelpers"
	"github.com/shortlink-org/shop/oms/pkg/uow"
	uowpg "github.com/shortlink-org/shop/oms/pkg/uow/postgres"
)

const uowMigration = `
CREATE SCHEMA IF NOT EXISTS oms;

CREATE TABLE IF NOT EXISTS oms.tx_test_entries (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL
);
`

func setupUoWTest(t *testing.T) (*uowpg.UoW, *testhelpers.PostgresContainer) {
	t.Helper()

	pc := testhelpers.SetupPostgresContainer(t)
	pc.RunMigrations(t, uowMigration)

	return uowpg.New(pc.Pool), pc
}

func TestUoW_CommitPersistsChanges(t *testing.T) {
	unitOfWork, pc := setupUoWTest(t)
	ctx := context.Background()

	txCtx, err := unitOfWork.Begin(ctx)
	require.NoError(t, err)

	tx := uow.FromContext(txCtx)
	require.NotNil(t, tx)

	_, err = tx.Exec(txCtx, `INSERT INTO oms.tx_test_entries (value) VALUES ($1)`, "persisted")
	require.NoError(t, err)

	err = unitOfWork.Commit(txCtx)
	require.NoError(t, err)

	var count int
	err = pc.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM oms.tx_test_entries WHERE value = $1`, "persisted").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUoW_RollbackDiscardsChanges(t *testing.T) {
	unitOfWork, pc := setupUoWTest(t)
	ctx := context.Background()

	txCtx, err := unitOfWork.Begin(ctx)
	require.NoError(t, err)

	tx := uow.FromContext(txCtx)
	require.NotNil(t, tx)

	_, err = tx.Exec(txCtx, `INSERT INTO oms.tx_test_entries (value) VALUES ($1)`, "rolled-back")
	require.NoError(t, err)

	err = unitOfWork.Rollback(txCtx)
	require.NoError(t, err)

	var count int
	err = pc.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM oms.tx_test_entries WHERE value = $1`, "rolled-back").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestUoW_RollbackAfterCommitIsNoop(t *testing.T) {
	unitOfWork, pc := setupUoWTest(t)
	ctx := context.Background()

	txCtx, err := unitOfWork.Begin(ctx)
	require.NoError(t, err)

	tx := uow.FromContext(txCtx)
	require.NotNil(t, tx)

	_, err = tx.Exec(txCtx, `INSERT INTO oms.tx_test_entries (value) VALUES ($1)`, "commit-then-rollback")
	require.NoError(t, err)

	err = unitOfWork.Commit(txCtx)
	require.NoError(t, err)

	err = unitOfWork.Rollback(txCtx)
	require.NoError(t, err)

	var count int
	err = pc.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM oms.tx_test_entries WHERE value = $1`, "commit-then-rollback").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
