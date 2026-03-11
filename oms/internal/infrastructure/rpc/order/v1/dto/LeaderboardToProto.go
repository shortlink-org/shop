package dto

import (
	"time"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	v1 "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/order/v1/model/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GoodsLeaderboardToProto(in *ports.GoodsLeaderboard) *v1.GoodsLeaderboard {
	if in == nil {
		return nil
	}

	entries := make([]*v1.LeaderboardEntry, 0, len(in.Entries))
	for _, entry := range in.Entries {
		entries = append(entries, &v1.LeaderboardEntry{
			MemberId: entry.MemberID.String(),
			Rank:     entry.Rank,
			Score:    entry.Score,
			Orders:   entry.Orders,
			Units:    entry.Units,
		})
	}

	return &v1.GoodsLeaderboard{
		Board:       string(in.Board),
		Window:      string(in.Window),
		GeneratedAt: timeToProto(in.GeneratedAt),
		Entries:     entries,
	}
}

func timeToProto(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}

	return timestamppb.New(value.UTC())
}
