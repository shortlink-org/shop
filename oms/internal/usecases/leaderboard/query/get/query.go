package get

import "github.com/shortlink-org/shop/oms/internal/domain/ports"

type Query struct {
	Board  ports.LeaderboardBoard
	Window ports.LeaderboardWindow
	Limit  int
}

func NewQuery(board ports.LeaderboardBoard, window ports.LeaderboardWindow, limit int) Query {
	return Query{
		Board:  board,
		Window: window,
		Limit:  limit,
	}
}
