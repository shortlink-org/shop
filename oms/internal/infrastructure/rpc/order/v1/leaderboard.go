package v1

import (
	"context"
	"fmt"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/dto"
	model "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	leaderboardget "github.com/shortlink-org/shop/oms/internal/usecases/leaderboard/query/get"
)

func (o *OrderRPC) GetLeaderboard(
	ctx context.Context,
	in *model.GetLeaderboardRequest,
) (*model.GetLeaderboardResponse, error) {
	board, err := ports.ParseLeaderboardBoard(in.GetBoard())
	if err != nil {
		return nil, fmt.Errorf("parse leaderboard board: %w", err)
	}

	window, err := ports.ParseLeaderboardWindow(in.GetWindow())
	if err != nil {
		return nil, fmt.Errorf("parse leaderboard window: %w", err)
	}

	result, err := o.leaderboardHandler.Handle(ctx, leaderboardget.NewQuery(board, window, int(in.GetLimit())))
	if err != nil {
		return nil, err
	}

	return &model.GetLeaderboardResponse{
		Leaderboard: dto.GoodsLeaderboardToProto(result),
	}, nil
}
