package dto

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	ordermodel "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/infrastructure/rpc/order/v1/model/v1"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func GoodsLeaderboardToService(in *ordermodel.GoodsLeaderboard) *servicepb.GoodsLeaderboard {
	if in == nil {
		return nil
	}

	entries := make([]*servicepb.LeaderboardEntry, 0, len(in.GetEntries()))
	for _, entry := range in.GetEntries() {
		entries = append(entries, &servicepb.LeaderboardEntry{
			MemberId: wrapperspb.String(entry.GetMemberId()),
			Rank:     wrapperspb.Int32(entry.GetRank()),
			Score:    wrapperspb.Double(entry.GetScore()),
			Orders:   wrapperspb.Int64(entry.GetOrders()),
			Units:    wrapperspb.Int64(entry.GetUnits()),
		})
	}

	return &servicepb.GoodsLeaderboard{
		Board:       wrapperspb.String(in.GetBoard()),
		Window:      wrapperspb.String(in.GetWindow()),
		GeneratedAt: cloneTimestamp(in.GetGeneratedAt()),
		Entries:     entries,
	}
}
