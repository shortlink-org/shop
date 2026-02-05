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

	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	orderrepo "github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/testhelpers"
	uowpg "github.com/shortlink-org/shop/oms/pkg/uow/postgres"
)

const orderMigration = `
CREATE SCHEMA IF NOT EXISTS oms;

CREATE TABLE IF NOT EXISTS oms.orders (
    id          UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_customer_id_idx ON oms.orders(customer_id);
CREATE INDEX IF NOT EXISTS orders_status_idx ON oms.orders(status);

CREATE TABLE IF NOT EXISTS oms.order_items (
    order_id  UUID NOT NULL REFERENCES oms.orders(id) ON DELETE CASCADE,
    good_id   UUID NOT NULL,
    quantity  INT NOT NULL CHECK (quantity > 0),
    price     DECIMAL(12,2) NOT NULL,
    PRIMARY KEY (order_id, good_id)
);
`

func setupOrderTest(t *testing.T) (*orderrepo.Store, *uowpg.UoW, *testhelpers.PostgresContainer) {
	t.Helper()

	pc := testhelpers.SetupPostgresContainer(t)
	pc.RunMigrations(t, orderMigration)

	store, err := orderrepo.New(context.Background(), pc.DB())
	require.NoError(t, err, "failed to create order repository")

	uow := uowpg.New(pc.Pool)

	return store, uow, pc
}

func createOrderWithItems(t *testing.T, customerID uuid.UUID, items order.Items) *order.OrderState {
	t.Helper()
	orderState := order.NewOrderState(customerID)
	err := orderState.CreateOrder(context.Background(), items)
	require.NoError(t, err)
	return orderState
}

func TestOrder_SaveAndLoad(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	items := order.Items{
		order.NewItem(uuid.New(), 2, decimal.NewFromFloat(19.99)),
		order.NewItem(uuid.New(), 1, decimal.NewFromFloat(49.99)),
	}

	orderState := createOrderWithItems(t, customerID, items)
	orderID := orderState.GetOrderID()

	// Save order within transaction
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)

	err = store.Save(txCtx, orderState)
	require.NoError(t, err)

	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load order in a new transaction
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx2)

	loaded, err := store.Load(txCtx2, orderID)
	require.NoError(t, err)

	// Verify loaded order
	assert.Equal(t, orderID, loaded.GetOrderID())
	assert.Equal(t, customerID, loaded.GetCustomerId())
	assert.Equal(t, order.OrderStatus_ORDER_STATUS_PROCESSING, loaded.GetStatus())
	assert.Equal(t, 1, loaded.GetVersion())

	loadedItems := loaded.GetItems()
	require.Len(t, loadedItems, 2)
}

func TestOrder_ListByCustomer(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	otherCustomerID := uuid.New()

	// Create multiple orders for the same customer
	order1 := createOrderWithItems(t, customerID, order.Items{
		order.NewItem(uuid.New(), 1, decimal.NewFromFloat(10.00)),
	})
	order2 := createOrderWithItems(t, customerID, order.Items{
		order.NewItem(uuid.New(), 2, decimal.NewFromFloat(20.00)),
	})
	order3 := createOrderWithItems(t, otherCustomerID, order.Items{
		order.NewItem(uuid.New(), 3, decimal.NewFromFloat(30.00)),
	})

	// Save all orders
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, order1)
	require.NoError(t, err)
	err = store.Save(txCtx, order2)
	require.NoError(t, err)
	err = store.Save(txCtx, order3)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// List orders for first customer
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx2)

	orders, err := store.ListByCustomer(txCtx2, customerID)
	require.NoError(t, err)

	assert.Len(t, orders, 2)
	for _, o := range orders {
		assert.Equal(t, customerID, o.GetCustomerId())
	}

	// List orders for second customer
	otherOrders, err := store.ListByCustomer(txCtx2, otherCustomerID)
	require.NoError(t, err)

	assert.Len(t, otherOrders, 1)
	assert.Equal(t, otherCustomerID, otherOrders[0].GetCustomerId())
}

func TestOrder_StatusTransition(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	items := order.Items{
		order.NewItem(uuid.New(), 1, decimal.NewFromFloat(100.00)),
	}

	orderState := createOrderWithItems(t, customerID, items)
	orderID := orderState.GetOrderID()

	// Save initial order (PROCESSING status after CreateOrder)
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, orderState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load and complete the order
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)

	loaded, err := store.Load(txCtx2, orderID)
	require.NoError(t, err)
	require.Equal(t, order.OrderStatus_ORDER_STATUS_PROCESSING, loaded.GetStatus())

	err = loaded.CompleteOrder()
	require.NoError(t, err)

	err = store.Save(txCtx2, loaded)
	require.NoError(t, err)
	err = uow.Commit(txCtx2)
	require.NoError(t, err)

	// Verify final status
	txCtx3, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx3)

	final, err := store.Load(txCtx3, orderID)
	require.NoError(t, err)

	assert.Equal(t, order.OrderStatus_ORDER_STATUS_COMPLETED, final.GetStatus())
	assert.Equal(t, 2, final.GetVersion())
}

func TestOrder_CancelOrder(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	items := order.Items{
		order.NewItem(uuid.New(), 1, decimal.NewFromFloat(50.00)),
	}

	orderState := createOrderWithItems(t, customerID, items)
	orderID := orderState.GetOrderID()

	// Save initial order
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, orderState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load and cancel the order
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)

	loaded, err := store.Load(txCtx2, orderID)
	require.NoError(t, err)

	err = loaded.CancelOrder()
	require.NoError(t, err)

	err = store.Save(txCtx2, loaded)
	require.NoError(t, err)
	err = uow.Commit(txCtx2)
	require.NoError(t, err)

	// Verify cancelled status
	txCtx3, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx3)

	final, err := store.Load(txCtx3, orderID)
	require.NoError(t, err)

	assert.Equal(t, order.OrderStatus_ORDER_STATUS_CANCELLED, final.GetStatus())
}

func TestOrder_OptimisticConcurrency(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	items := order.Items{
		order.NewItem(uuid.New(), 1, decimal.NewFromFloat(10.00)),
	}

	orderState := createOrderWithItems(t, customerID, items)
	orderID := orderState.GetOrderID()

	// Save initial order
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, orderState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load order in two separate transactions
	txCtx1, err := uow.Begin(ctx)
	require.NoError(t, err)
	order1, err := store.Load(txCtx1, orderID)
	require.NoError(t, err)

	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	order2, err := store.Load(txCtx2, orderID)
	require.NoError(t, err)

	// First transaction completes the order
	err = order1.CompleteOrder()
	require.NoError(t, err)
	err = store.Save(txCtx1, order1)
	require.NoError(t, err)
	err = uow.Commit(txCtx1)
	require.NoError(t, err)

	// Second transaction tries to cancel with stale version - should fail
	err = order2.CancelOrder()
	require.NoError(t, err)
	err = store.Save(txCtx2, order2)

	assert.True(t, errors.Is(err, ports.ErrVersionConflict), "expected ErrVersionConflict, got: %v", err)

	uow.Rollback(txCtx2)
}

func TestOrder_LoadNotFound(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	nonExistentID := uuid.New()

	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx)

	_, err = store.Load(txCtx, nonExistentID)

	assert.True(t, errors.Is(err, ports.ErrNotFound), "expected ErrNotFound, got: %v", err)
}

func TestOrder_ListByCustomerEmpty(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()

	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx)

	orders, err := store.ListByCustomer(txCtx, customerID)
	require.NoError(t, err)

	assert.Empty(t, orders)
}

func TestOrder_MultipleItemsPreserved(t *testing.T) {
	store, uow, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	goodIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	items := order.Items{
		order.NewItem(goodIDs[0], 5, decimal.NewFromFloat(15.00)),
		order.NewItem(goodIDs[1], 2, decimal.NewFromFloat(25.50)),
		order.NewItem(goodIDs[2], 1, decimal.NewFromFloat(99.99)),
	}

	orderState := createOrderWithItems(t, customerID, items)
	orderID := orderState.GetOrderID()

	// Save
	txCtx, err := uow.Begin(ctx)
	require.NoError(t, err)
	err = store.Save(txCtx, orderState)
	require.NoError(t, err)
	err = uow.Commit(txCtx)
	require.NoError(t, err)

	// Load and verify
	txCtx2, err := uow.Begin(ctx)
	require.NoError(t, err)
	defer uow.Rollback(txCtx2)

	loaded, err := store.Load(txCtx2, orderID)
	require.NoError(t, err)

	loadedItems := loaded.GetItems()
	require.Len(t, loadedItems, 3)

	// Create a map for easy lookup
	loadedMap := make(map[uuid.UUID]order.Item)
	for _, item := range loadedItems {
		loadedMap[item.GetGoodId()] = item
	}

	// Verify each item
	item0, exists := loadedMap[goodIDs[0]]
	require.True(t, exists)
	assert.Equal(t, int32(5), item0.GetQuantity())
	assert.True(t, item0.GetPrice().Equal(decimal.NewFromFloat(15.00)))

	item1, exists := loadedMap[goodIDs[1]]
	require.True(t, exists)
	assert.Equal(t, int32(2), item1.GetQuantity())
	assert.True(t, item1.GetPrice().Equal(decimal.NewFromFloat(25.50)))

	item2, exists := loadedMap[goodIDs[2]]
	require.True(t, exists)
	assert.Equal(t, int32(1), item2.GetQuantity())
	assert.True(t, item2.GetPrice().Equal(decimal.NewFromFloat(99.99)))
}
