package get

import (
	"context"
	"time"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

type Result = *ports.GoodsLeaderboard

type Handler struct {
	repo ports.LeaderboardRepository
}

func NewHandler(repo ports.LeaderboardRepository) (*Handler, error) {
	return &Handler{repo: repo}, nil
}

func (h *Handler) Handle(ctx context.Context, q Query) (Result, error) {
	return h.repo.GetGoods(ctx, q.Board, q.Window, q.Limit, time.Now().UTC())
}
