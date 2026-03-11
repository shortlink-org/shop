package get

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

func TestHandlerHandleDelegatesToRepository(t *testing.T) {
	t.Parallel()

	wantGeneratedAt := time.Date(2026, time.March, 11, 14, 0, 0, 0, time.UTC)
	wantResult := &ports.GoodsLeaderboard{
		Board:       ports.LeaderboardBoardGoodsGMV,
		Window:      ports.LeaderboardWindowWeek,
		GeneratedAt: &wantGeneratedAt,
		Entries: []ports.LeaderboardEntry{
			{
				MemberID: uuid.New(),
				Rank:     1,
				Score:    123.45,
				Orders:   3,
				Units:    8,
			},
		},
	}

	repo := &leaderboardRepositoryStub{
		getGoodsResult: wantResult,
	}

	handler, err := NewHandler(repo)
	require.NoError(t, err)

	start := time.Now().UTC()
	result, err := handler.Handle(context.Background(), NewQuery(
		ports.LeaderboardBoardGoodsGMV,
		ports.LeaderboardWindowWeek,
		5,
	))
	end := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, wantResult, result)
	require.Equal(t, ports.LeaderboardBoardGoodsGMV, repo.getGoodsBoard)
	require.Equal(t, ports.LeaderboardWindowWeek, repo.getGoodsWindow)
	require.Equal(t, 5, repo.getGoodsLimit)
	require.False(t, repo.getGoodsNow.Before(start))
	require.False(t, repo.getGoodsNow.After(end))
}

type leaderboardRepositoryStub struct {
	getGoodsBoard  ports.LeaderboardBoard
	getGoodsWindow ports.LeaderboardWindow
	getGoodsLimit  int
	getGoodsNow    time.Time
	getGoodsResult *ports.GoodsLeaderboard
	getGoodsErr    error
}

func (s *leaderboardRepositoryStub) ApplyOrderCompleted(context.Context, ports.CompletedOrderLeaderboardSnapshot) (bool, error) {
	return false, nil
}

func (s *leaderboardRepositoryStub) GetGoods(
	_ context.Context,
	board ports.LeaderboardBoard,
	window ports.LeaderboardWindow,
	limit int,
	now time.Time,
) (*ports.GoodsLeaderboard, error) {
	s.getGoodsBoard = board
	s.getGoodsWindow = window
	s.getGoodsLimit = limit
	s.getGoodsNow = now

	return s.getGoodsResult, s.getGoodsErr
}
