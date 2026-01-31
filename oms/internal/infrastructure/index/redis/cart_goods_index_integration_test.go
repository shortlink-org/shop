//go:build integration

package redis_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redisindex "github.com/shortlink-org/shop/oms/internal/infrastructure/index/redis"
	"github.com/shortlink-org/shop/oms/internal/testhelpers"
)

func setupRedisTest(t *testing.T) (*redisindex.CartGoodsIndex, *testhelpers.RedisContainer) {
	t.Helper()

	rc := testhelpers.SetupRedisContainer(t)
	index := redisindex.New(rc.Client)

	return index, rc
}

func TestIndex_AddAndGetCustomers(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	goodID := uuid.New()
	customer1 := uuid.New()
	customer2 := uuid.New()
	customer3 := uuid.New()

	// Add good to multiple customers' carts
	err := index.AddGoodToCart(ctx, goodID, customer1)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, goodID, customer2)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, goodID, customer3)
	require.NoError(t, err)

	// Get customers with the good
	customers, err := index.GetCustomersWithGood(ctx, goodID)
	require.NoError(t, err)

	assert.Len(t, customers, 3)

	// Verify all customers are present
	customerMap := make(map[uuid.UUID]bool)
	for _, c := range customers {
		customerMap[c] = true
	}

	assert.True(t, customerMap[customer1])
	assert.True(t, customerMap[customer2])
	assert.True(t, customerMap[customer3])
}

func TestIndex_AddMultipleGoods(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	good1 := uuid.New()
	good2 := uuid.New()
	good3 := uuid.New()

	// Add multiple goods to the same customer's cart
	err := index.AddGoodToCart(ctx, good1, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, good2, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, good3, customerID)
	require.NoError(t, err)

	// Verify each good maps to the customer
	for _, goodID := range []uuid.UUID{good1, good2, good3} {
		customers, err := index.GetCustomersWithGood(ctx, goodID)
		require.NoError(t, err)
		require.Len(t, customers, 1)
		assert.Equal(t, customerID, customers[0])
	}
}

func TestIndex_RemoveGood(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	goodID := uuid.New()
	customer1 := uuid.New()
	customer2 := uuid.New()

	// Add good to two customers
	err := index.AddGoodToCart(ctx, goodID, customer1)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, goodID, customer2)
	require.NoError(t, err)

	// Remove from first customer
	err = index.RemoveGoodFromCart(ctx, goodID, customer1)
	require.NoError(t, err)

	// Verify only second customer remains
	customers, err := index.GetCustomersWithGood(ctx, goodID)
	require.NoError(t, err)

	assert.Len(t, customers, 1)
	assert.Equal(t, customer2, customers[0])
}

func TestIndex_RemoveNonExistent(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	goodID := uuid.New()
	customerID := uuid.New()

	// Remove from non-existent good should not error
	err := index.RemoveGoodFromCart(ctx, goodID, customerID)
	require.NoError(t, err)
}

func TestIndex_ClearCart(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	good1 := uuid.New()
	good2 := uuid.New()
	good3 := uuid.New()

	// Add multiple goods to customer's cart
	err := index.AddGoodToCart(ctx, good1, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, good2, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, good3, customerID)
	require.NoError(t, err)

	// Clear the cart
	err = index.ClearCart(ctx, customerID)
	require.NoError(t, err)

	// Verify all goods are removed from this customer
	for _, goodID := range []uuid.UUID{good1, good2, good3} {
		customers, err := index.GetCustomersWithGood(ctx, goodID)
		require.NoError(t, err)
		assert.Empty(t, customers)
	}
}

func TestIndex_ClearCartPreservesOtherCustomers(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	customer1 := uuid.New()
	customer2 := uuid.New()
	sharedGood := uuid.New()

	// Both customers have the same good
	err := index.AddGoodToCart(ctx, sharedGood, customer1)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, sharedGood, customer2)
	require.NoError(t, err)

	// Clear first customer's cart
	err = index.ClearCart(ctx, customer1)
	require.NoError(t, err)

	// Verify second customer still has the good
	customers, err := index.GetCustomersWithGood(ctx, sharedGood)
	require.NoError(t, err)

	assert.Len(t, customers, 1)
	assert.Equal(t, customer2, customers[0])
}

func TestIndex_GetCustomersWithNonExistentGood(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	nonExistentGood := uuid.New()

	customers, err := index.GetCustomersWithGood(ctx, nonExistentGood)
	require.NoError(t, err)

	assert.Empty(t, customers)
}

func TestIndex_BidirectionalConsistency(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	good1 := uuid.New()
	good2 := uuid.New()

	// Add goods
	err := index.AddGoodToCart(ctx, good1, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, good2, customerID)
	require.NoError(t, err)

	// Remove one good
	err = index.RemoveGoodFromCart(ctx, good1, customerID)
	require.NoError(t, err)

	// Verify good1 no longer maps to customer
	customers1, err := index.GetCustomersWithGood(ctx, good1)
	require.NoError(t, err)
	assert.Empty(t, customers1)

	// Verify good2 still maps to customer
	customers2, err := index.GetCustomersWithGood(ctx, good2)
	require.NoError(t, err)
	assert.Len(t, customers2, 1)
	assert.Equal(t, customerID, customers2[0])

	// Clear cart and verify all cleaned up
	err = index.ClearCart(ctx, customerID)
	require.NoError(t, err)

	customers2AfterClear, err := index.GetCustomersWithGood(ctx, good2)
	require.NoError(t, err)
	assert.Empty(t, customers2AfterClear)
}

func TestIndex_DuplicateAdd(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	goodID := uuid.New()
	customerID := uuid.New()

	// Add same good/customer multiple times
	err := index.AddGoodToCart(ctx, goodID, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, goodID, customerID)
	require.NoError(t, err)

	err = index.AddGoodToCart(ctx, goodID, customerID)
	require.NoError(t, err)

	// Should still only have one entry (idempotent)
	customers, err := index.GetCustomersWithGood(ctx, goodID)
	require.NoError(t, err)

	assert.Len(t, customers, 1)
	assert.Equal(t, customerID, customers[0])
}

func TestIndex_ClearEmptyCart(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	customerID := uuid.New()

	// Clear cart that doesn't exist should not error
	err := index.ClearCart(ctx, customerID)
	require.NoError(t, err)
}

func TestIndex_LargeNumberOfCustomers(t *testing.T) {
	index, _ := setupRedisTest(t)
	ctx := context.Background()

	goodID := uuid.New()
	numCustomers := 100

	// Add good to many customers
	expectedCustomers := make(map[uuid.UUID]bool)
	for i := 0; i < numCustomers; i++ {
		customerID := uuid.New()
		expectedCustomers[customerID] = true

		err := index.AddGoodToCart(ctx, goodID, customerID)
		require.NoError(t, err)
	}

	// Verify all customers are returned
	customers, err := index.GetCustomersWithGood(ctx, goodID)
	require.NoError(t, err)

	assert.Len(t, customers, numCustomers)

	for _, c := range customers {
		assert.True(t, expectedCustomers[c], "unexpected customer: %s", c)
	}
}
