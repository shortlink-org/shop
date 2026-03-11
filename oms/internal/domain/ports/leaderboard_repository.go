package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrUnsupportedLeaderboardBoard  = errors.New("unsupported leaderboard board")
	ErrUnsupportedLeaderboardWindow = errors.New("unsupported leaderboard window")
)

type LeaderboardBoard string

const (
	LeaderboardBoardGoodsGMV    LeaderboardBoard = "GOODS_GMV"
	LeaderboardBoardGoodsOrders LeaderboardBoard = "GOODS_ORDERS"
	LeaderboardBoardGoodsUnits  LeaderboardBoard = "GOODS_UNITS"
)

func ParseLeaderboardBoard(raw string) (LeaderboardBoard, error) {
	switch LeaderboardBoard(raw) {
	case LeaderboardBoardGoodsGMV, LeaderboardBoardGoodsOrders, LeaderboardBoardGoodsUnits:
		return LeaderboardBoard(raw), nil
	default:
		return "", ErrUnsupportedLeaderboardBoard
	}
}

type LeaderboardWindow string

const (
	LeaderboardWindowDay   LeaderboardWindow = "DAY"
	LeaderboardWindowWeek  LeaderboardWindow = "WEEK"
	LeaderboardWindowMonth LeaderboardWindow = "MONTH"
)

func ParseLeaderboardWindow(raw string) (LeaderboardWindow, error) {
	switch LeaderboardWindow(raw) {
	case LeaderboardWindowDay, LeaderboardWindowWeek, LeaderboardWindowMonth:
		return LeaderboardWindow(raw), nil
	default:
		return "", ErrUnsupportedLeaderboardWindow
	}
}

type LeaderboardOrderItem struct {
	GoodID   uuid.UUID
	Quantity int32
	Price    decimal.Decimal
}

type CompletedOrderLeaderboardSnapshot struct {
	OrderID          uuid.UUID
	AggregateVersion int32
	CompletedAt      time.Time
	Items            []LeaderboardOrderItem
}

type LeaderboardEntry struct {
	MemberID uuid.UUID
	Rank     int32
	Score    float64
	Orders   int64
	Units    int64
}

type GoodsLeaderboard struct {
	Board       LeaderboardBoard
	Window      LeaderboardWindow
	GeneratedAt *time.Time
	Entries     []LeaderboardEntry
}

// LeaderboardRepository manages leaderboard projections and read models.
//
//nolint:iface // port interface used by use cases and DI
type LeaderboardRepository interface {
	ApplyOrderCompleted(ctx context.Context, snapshot CompletedOrderLeaderboardSnapshot) (bool, error)
	GetGoods(ctx context.Context, board LeaderboardBoard, window LeaderboardWindow, limit int, now time.Time) (*GoodsLeaderboard, error)
}
