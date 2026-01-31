//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shop/oms/internal/boundary/ports"
	cart "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	cartrepo "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/cart"
	"github.com/shortlink-org/shop/oms/internal/testhelpers"
	uowpg "github.com/shortlink-org/shop/oms/pkg/uow/postgres"
)

const cartMigration = `
CREATE SCHEMA IF NOT EXISTS oms;

CREATE TABLE IF NOT EXISTS oms.carts (
    customer_id UUID PRIMARY KEY,
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oms.cart_items (
    cart_id   UUID NOT NULL REFERENCES oms.carts(customer_id) ON DELETE CASCADE,
    good_id   UUID NOT NULL,
    quantity  INT NOT NULL CHECK (quantity > 0),
    price     DECIMAL(12,2) NOT NULL,
    discount  DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    PRIMARY KEY (cart_id, good_id)
);
`

func setupCartTest(t *testing.T) (*cartrepo.Store, *uowpg.UoW, *testhelpers.PostgresContainer) {
	t.Helper()

	pc := testhelpers.SetupPostgresContainer(t)
	pc.RunMigrations(t, cartMigration)

	store, err := cartrepo.New(context.Background(), pc.DB())
	require.NoError(t, err, "failed to create cart repository")

	uow := uowpg.New(pc.Pool)

	return store, uow, pc
}

func mustNewItem(t *testing.T, goodID uuid.UUID, quantity int32, price, discount decimal.Decimal) itemv1.Item {
	t.Helper()
	item, err := itemv1.NewItemWithPricing(goodID, quantity, price, discount, decimal.Zero)
	require.NoError(t, err)
	return item
}

func TestCart_SaveAndLoad(t *testing.T) {
	store, uow, _ := setupCartTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	goodID := uuid.New()

	// Create cart with items
	cartState := cart.New(customerID)
	item := mustNewItem(t, goodID, 2, decimal.NewFromFloat(19.99), decimal.NewFromFloat(2.00))
	err := cartState.AddItem(item)
	require.NoError(t, err)

	// Save cart within transaction
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)

	err = store.Save(txCtx, cartState)
	require.NoError(t, err)

	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load cart in a new transaction
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx2)

	loaded, err := store.Load(txCtx2, customerID)
	require.NoError(t, err)

	// Verify loaded cart
	assert.Equal(t, customerID, loaded.GetCustomerId())
	assert.Equal(t, 1, loaded.GetVersion())

	items := loaded.GetItems()
	require.Len(t, items, 1)
	assert.Equal(t, goodID, items[0].GetGoodId())
	assert.Equal(t, int32(2), items[0].GetQuantity())
	assert.True(t, items[0].GetPrice().Equal(decimal.NewFromFloat(19.99)))
	assert.True(t, items[0].GetDiscount().Equal(decimal.NewFromFloat(2.00)))
}

func TestCart_UpdateExistingCart(t *testing.T) {
	store, uow, _ := setupCartTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	goodID1 := uuid.New()
	goodID2 := uuid.New()

	// Create and save initial cart
	cartState := cart.New(customerID)
	item1 := mustNewItem(t, goodID1, 1, decimal.NewFromFloat(10.00), decimal.Zero)
	err := cartState.AddItem(item1)
	require.NoError(t, err)

	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, cartState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load cart and add another item
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)

	loaded, err := store.Load(txCtx2, customerID)
	require.NoError(t, err)

	item2 := mustNewItem(t, goodID2, 3, decimal.NewFromFloat(5.50), decimal.NewFromFloat(0.50))
	err = loaded.AddItem(item2)
	require.NoError(t, err)

	err = store.Save(txCtx2, loaded)
	require.NoError(t, err)
	err = uow.Commit(txCtx2)
	require.NoError(t, err)

	// Verify updated cart
	txCtx3, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx3)

	final, err := store.Load(txCtx3, customerID)
	require.NoError(t, err)

	assert.Equal(t, 2, final.GetVersion())
	assert.Len(t, final.GetItems(), 2)
}

func TestCart_OptimisticConcurrency(t *testing.T) {
	store, uow, _ := setupCartTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	goodID := uuid.New()

	// Create and save initial cart
	cartState := cart.New(customerID)
	item := mustNewItem(t, goodID, 1, decimal.NewFromFloat(10.00), decimal.Zero)
	err := cartState.AddItem(item)
	require.NoError(t, err)

	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, cartState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load cart in two separate transactions (simulating concurrent access)
	txCtx1, err := uow.Begin(ctx)
	require.NoError(t, err)
	cart1, err := store.Load(txCtx1, customerID)
	require.NoError(t, err)

	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	cart2, err := store.Load(txCtx2, customerID)
	require.NoError(t, err)

	// First transaction updates and commits successfully
	newItem1 := mustNewItem(t, uuid.New(), 2, decimal.NewFromFloat(20.00), decimal.Zero)
	err = cart1.AddItem(newItem1)
	require.NoError(t, err)
	err = store.Save(txCtx1, cart1)
	require.NoError(t, err)
	err = uow.Commit(txCtx1)
	require.NoError(t, err)

	// Second transaction tries to update with stale version - should fail
	newItem2 := mustNewItem(t, uuid.New(), 3, decimal.NewFromFloat(30.00), decimal.Zero)
	err = cart2.AddItem(newItem2)
	require.NoError(t, err)
	err = store.Save(txCtx2, cart2)

	assert.True(t, errors.Is(err, ports.ErrVersionConflict), "expected ErrVersionConflict, got: %v", err)

	// Rollback second transaction
	uow.Rollback(txCtx2)
}

func TestCart_LoadNotFound(t *testing.T) {
	store, uow, _ := setupCartTest(t)
	ctx := context.Background()

	nonExistentID := uuid.New()

	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx)

	_, err = store.Load(txCtx, nonExistentID)

	assert.True(t, errors.Is(err, ports.ErrNotFound), "expected ErrNotFound, got: %v", err)
}

func TestCart_ClearItems(t *testing.T) {
	store, uow, _ := setupCartTest(t)
	ctx := context.Background()

	customerID := uuid.New()

	// Create cart with multiple items
	cartState := cart.New(customerID)
	for i := 0; i < 3; i++ {
		item := mustNewItem(t, uuid.New(), int32(i+1), decimal.NewFromFloat(float64(i+1)*10), decimal.Zero)
		err := cartState.AddItem(item)
		require.NoError(t, err)
	}

	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, cartState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load and clear cart
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)

	loaded, err := store.Load(txCtx2, customerID)
	require.NoError(t, err)
	require.Len(t, loaded.GetItems(), 3)

	loaded.Reset()

	err = store.Save(txCtx2, loaded)
	require.NoError(t, err)
	err = uow.Commit(txCtx2)
	require.NoError(t, err)

	// Verify cart is empty
	txCtx3, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx3)

	final, err := store.Load(txCtx3, customerID)
	require.NoError(t, err)

	assert.Empty(t, final.GetItems())
	assert.Equal(t, 2, final.GetVersion())
}

func TestCart_MultipleItemsWithPricing(t *testing.T) {
	store, uow, _ := setupCartTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	cartState := cart.New(customerID)

	items := []struct {
		goodID   uuid.UUID
		quantity int32
		price    decimal.Decimal
		discount decimal.Decimal
	}{
		{uuid.New(), 1, decimal.NewFromFloat(100.00), decimal.NewFromFloat(10.00)},
		{uuid.New(), 5, decimal.NewFromFloat(25.50), decimal.NewFromFloat(5.50)},
		{uuid.New(), 2, decimal.NewFromFloat(49.99), decimal.Zero},
	}

	for _, it := range items {
		item := mustNewItem(t, it.goodID, it.quantity, it.price, it.discount)
		err := cartState.AddItem(item)
		require.NoError(t, err)
	}

	// Save
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, cartState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load and verify
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx2)

	loaded, err := store.Load(txCtx2, customerID)
	require.NoError(t, err)

	loadedItems := loaded.GetItems()
	require.Len(t, loadedItems, 3)

	// Create a map for easy lookup
	loadedMap := make(map[uuid.UUID]itemv1.Item)
	for _, item := range loadedItems {
		loadedMap[item.GetGoodId()] = item
	}

	for _, expected := range items {
		actual, exists := loadedMap[expected.goodID]
		require.True(t, exists, "item %s not found", expected.goodID)
		assert.Equal(t, expected.quantity, actual.GetQuantity())
		assert.True(t, expected.price.Equal(actual.GetPrice()))
		assert.True(t, expected.discount.Equal(actual.GetDiscount()))
	}
}
