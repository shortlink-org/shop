package leaderboard

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

func TestStoreApplyOrderCompletedAndGetGoodsRankings(t *testing.T) {
	t.Parallel()

	store, cleanup := newTestStore(t)
	defer cleanup()

	goodA := uuid.New()
	goodB := uuid.New()
	completedAt1 := time.Date(2026, time.March, 11, 10, 0, 0, 0, time.UTC)
	completedAt2 := completedAt1.Add(2 * time.Hour)

	applied, err := store.ApplyOrderCompleted(context.Background(), ports.CompletedOrderLeaderboardSnapshot{
		OrderID:          uuid.New(),
		AggregateVersion: 1,
		CompletedAt:      completedAt1,
		Items: []ports.LeaderboardOrderItem{
			{GoodID: goodA, Quantity: 2, Price: decimal.NewFromInt(10)},
			{GoodID: goodB, Quantity: 1, Price: decimal.NewFromFloat(5.5)},
		},
	})
	require.NoError(t, err)
	require.True(t, applied)

	applied, err = store.ApplyOrderCompleted(context.Background(), ports.CompletedOrderLeaderboardSnapshot{
		OrderID:          uuid.New(),
		AggregateVersion: 1,
		CompletedAt:      completedAt2,
		Items: []ports.LeaderboardOrderItem{
			{GoodID: goodB, Quantity: 3, Price: decimal.NewFromInt(20)},
		},
	})
	require.NoError(t, err)
	require.True(t, applied)

	gmvBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowDay,
		10,
		completedAt2,
	)
	require.NoError(t, err)
	require.NotNil(t, gmvBoard.GeneratedAt)
	require.True(t, gmvBoard.GeneratedAt.Equal(completedAt2))
	require.Len(t, gmvBoard.Entries, 2)
	require.Equal(t, goodB, gmvBoard.Entries[0].MemberID)
	require.InDelta(t, 65.5, gmvBoard.Entries[0].Score, 0.001)
	require.EqualValues(t, 2, gmvBoard.Entries[0].Orders)
	require.EqualValues(t, 4, gmvBoard.Entries[0].Units)
	require.Equal(t, goodA, gmvBoard.Entries[1].MemberID)
	require.InDelta(t, 20.0, gmvBoard.Entries[1].Score, 0.001)

	ordersBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsOrders,
		ports.LeaderboardWindowDay,
		10,
		completedAt2,
	)
	require.NoError(t, err)
	require.Len(t, ordersBoard.Entries, 2)
	require.Equal(t, goodB, ordersBoard.Entries[0].MemberID)
	require.InDelta(t, 2.0, ordersBoard.Entries[0].Score, 0.001)

	unitsBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsUnits,
		ports.LeaderboardWindowDay,
		10,
		completedAt2,
	)
	require.NoError(t, err)
	require.Len(t, unitsBoard.Entries, 2)
	require.Equal(t, goodB, unitsBoard.Entries[0].MemberID)
	require.InDelta(t, 4.0, unitsBoard.Entries[0].Score, 0.001)

	weekBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowWeek,
		10,
		completedAt2,
	)
	require.NoError(t, err)
	require.Len(t, weekBoard.Entries, 2)
	require.Equal(t, goodB, weekBoard.Entries[0].MemberID)

	monthBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowMonth,
		10,
		completedAt2,
	)
	require.NoError(t, err)
	require.Len(t, monthBoard.Entries, 2)
	require.Equal(t, goodB, monthBoard.Entries[0].MemberID)
}

func TestStoreApplyOrderCompletedIsIdempotentAndWindowed(t *testing.T) {
	t.Parallel()

	store, cleanup := newTestStore(t)
	defer cleanup()

	goodID := uuid.New()
	orderID := uuid.New()
	completedAt := time.Date(2026, time.March, 11, 16, 0, 0, 0, time.UTC)

	snapshot := ports.CompletedOrderLeaderboardSnapshot{
		OrderID:          orderID,
		AggregateVersion: 7,
		CompletedAt:      completedAt,
		Items: []ports.LeaderboardOrderItem{
			{GoodID: goodID, Quantity: 2, Price: decimal.NewFromInt(10)},
		},
	}

	applied, err := store.ApplyOrderCompleted(context.Background(), snapshot)
	require.NoError(t, err)
	require.True(t, applied)

	applied, err = store.ApplyOrderCompleted(context.Background(), snapshot)
	require.NoError(t, err)
	require.False(t, applied)

	sameDay, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowDay,
		10,
		completedAt,
	)
	require.NoError(t, err)
	require.Len(t, sameDay.Entries, 1)
	require.InDelta(t, 20.0, sameDay.Entries[0].Score, 0.001)

	nextDay := completedAt.Add(24 * time.Hour)
	emptyDay, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowDay,
		10,
		nextDay,
	)
	require.NoError(t, err)
	require.Empty(t, emptyDay.Entries)

	weekBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowWeek,
		10,
		nextDay,
	)
	require.NoError(t, err)
	require.Len(t, weekBoard.Entries, 1)
	require.InDelta(t, 20.0, weekBoard.Entries[0].Score, 0.001)

	monthBoard, err := store.GetGoods(
		context.Background(),
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowMonth,
		10,
		time.Date(2026, time.March, 31, 10, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	require.Len(t, monthBoard.Entries, 1)
	require.InDelta(t, 20.0, monthBoard.Entries[0].Score, 0.001)
}

func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	mr := miniredis.RunT(t)
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return New(client), cleanup
}
