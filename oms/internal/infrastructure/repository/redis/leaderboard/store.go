package leaderboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/shopspring/decimal"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

const (
	keyPrefix = "shop:lb:v1:{goods}"

	dedupeRetention = 365 * 24 * time.Hour
	dayRetention    = 45 * 24 * time.Hour
	weekRetention   = 400 * 24 * time.Hour
	monthRetention  = 800 * 24 * time.Hour
)

var applyCompletedOrderScript = rueidis.NewLuaScriptNoShaRetryable(`
if redis.call('EXISTS', KEYS[1]) == 1 then
  return 0
end

redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])

local occurred_at = ARGV[3]
local day_ttl = tonumber(ARGV[4])
local week_ttl = tonumber(ARGV[5])
local month_ttl = tonumber(ARGV[6])
local item_count = tonumber(ARGV[7])
local idx = 8

for _ = 1, item_count do
  local member = ARGV[idx]
  local gmv = tonumber(ARGV[idx + 1])
  local orders = tonumber(ARGV[idx + 2])
  local units = tonumber(ARGV[idx + 3])
  idx = idx + 4

  redis.call('ZINCRBY', KEYS[5], gmv, member)
  redis.call('ZINCRBY', KEYS[6], gmv, member)
  redis.call('ZINCRBY', KEYS[7], gmv, member)

  redis.call('ZINCRBY', KEYS[8], orders, member)
  redis.call('ZINCRBY', KEYS[9], orders, member)
  redis.call('ZINCRBY', KEYS[10], orders, member)

  redis.call('ZINCRBY', KEYS[11], units, member)
  redis.call('ZINCRBY', KEYS[12], units, member)
  redis.call('ZINCRBY', KEYS[13], units, member)
end

redis.call('SET', KEYS[2], occurred_at, 'EX', day_ttl)
redis.call('SET', KEYS[3], occurred_at, 'EX', week_ttl)
redis.call('SET', KEYS[4], occurred_at, 'EX', month_ttl)

for _, key in ipairs({KEYS[5], KEYS[8], KEYS[11]}) do
  redis.call('EXPIRE', key, day_ttl)
end
for _, key in ipairs({KEYS[6], KEYS[9], KEYS[12]}) do
  redis.call('EXPIRE', key, week_ttl)
end
for _, key in ipairs({KEYS[7], KEYS[10], KEYS[13]}) do
  redis.call('EXPIRE', key, month_ttl)
end

return 1
`)

type Store struct {
	client rueidis.Client
}

func New(client rueidis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) ApplyOrderCompleted(
	ctx context.Context,
	snapshot ports.CompletedOrderLeaderboardSnapshot,
) (bool, error) {
	completedAt := snapshot.CompletedAt.UTC()
	dayBucket := bucketForWindow(ports.LeaderboardWindowDay, completedAt)
	weekBucket := bucketForWindow(ports.LeaderboardWindowWeek, completedAt)
	monthBucket := bucketForWindow(ports.LeaderboardWindowMonth, completedAt)

	aggregated := aggregateOrderItems(snapshot.Items)
	if len(aggregated) == 0 {
		return false, nil
	}

	keys := []string{
		processedEventKey(snapshot.OrderID, snapshot.AggregateVersion),
		generatedAtKey(ports.LeaderboardWindowDay, dayBucket),
		generatedAtKey(ports.LeaderboardWindowWeek, weekBucket),
		generatedAtKey(ports.LeaderboardWindowMonth, monthBucket),
		scoreKey(metricGMV, ports.LeaderboardWindowDay, dayBucket),
		scoreKey(metricGMV, ports.LeaderboardWindowWeek, weekBucket),
		scoreKey(metricGMV, ports.LeaderboardWindowMonth, monthBucket),
		scoreKey(metricOrders, ports.LeaderboardWindowDay, dayBucket),
		scoreKey(metricOrders, ports.LeaderboardWindowWeek, weekBucket),
		scoreKey(metricOrders, ports.LeaderboardWindowMonth, monthBucket),
		scoreKey(metricUnits, ports.LeaderboardWindowDay, dayBucket),
		scoreKey(metricUnits, ports.LeaderboardWindowWeek, weekBucket),
		scoreKey(metricUnits, ports.LeaderboardWindowMonth, monthBucket),
	}

	args := []string{
		completedAt.Format(time.RFC3339Nano),
		strconv.FormatInt(int64(dedupeRetention/time.Second), 10),
		completedAt.Format(time.RFC3339Nano),
		strconv.FormatInt(int64(dayRetention/time.Second), 10),
		strconv.FormatInt(int64(weekRetention/time.Second), 10),
		strconv.FormatInt(int64(monthRetention/time.Second), 10),
		strconv.Itoa(len(aggregated)),
	}

	for _, item := range aggregated {
		args = append(
			args,
			item.goodID.String(),
			item.gmv.StringFixedBank(2),
			strconv.FormatInt(item.orders, 10),
			strconv.FormatInt(item.units, 10),
		)
	}

	result, err := applyCompletedOrderScript.Exec(ctx, s.client, keys, args).AsInt64()
	if err != nil {
		return false, fmt.Errorf("apply leaderboard projection: %w", err)
	}

	return result == 1, nil
}

func (s *Store) GetGoods(
	ctx context.Context,
	board ports.LeaderboardBoard,
	window ports.LeaderboardWindow,
	limit int,
	now time.Time,
) (*ports.GoodsLeaderboard, error) {
	if limit <= 0 {
		limit = 7
	}

	bucket := bucketForWindow(window, now.UTC())
	mainMetric, err := metricForBoard(board)
	if err != nil {
		return nil, err
	}

	mainKey := scoreKey(mainMetric, window, bucket)
	ordersKey := scoreKey(metricOrders, window, bucket)
	unitsKey := scoreKey(metricUnits, window, bucket)

	rangeResp, err := s.client.Do(
		ctx,
		s.client.B().Zrevrange().Key(mainKey).Start(0).Stop(int64(limit-1)).Withscores().Build(),
	).ToArray()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return &ports.GoodsLeaderboard{Board: board, Window: window, Entries: []ports.LeaderboardEntry{}}, nil
		}

		return nil, fmt.Errorf("get leaderboard range: %w", err)
	}

	entries := make([]ports.LeaderboardEntry, 0, len(rangeResp)/2)
	memberIDs := make([]string, 0, len(rangeResp)/2)
	for i := 0; i+1 < len(rangeResp); i += 2 {
		memberID, err := rangeResp[i].ToString()
		if err != nil {
			return nil, fmt.Errorf("read leaderboard member id: %w", err)
		}

		score, err := rangeResp[i+1].AsFloat64()
		if err != nil {
			return nil, fmt.Errorf("read leaderboard score: %w", err)
		}

		goodID, err := uuid.Parse(memberID)
		if err != nil {
			return nil, fmt.Errorf("parse leaderboard member id: %w", err)
		}

		memberIDs = append(memberIDs, memberID)
		entries = append(entries, ports.LeaderboardEntry{
			MemberID: goodID,
			Rank:     int32(len(entries) + 1),
			Score:    score,
		})
	}

	if len(entries) > 0 {
		if err := s.fillMetricCounts(ctx, ordersKey, memberIDs, entries, metricOrders); err != nil {
			return nil, err
		}
		if err := s.fillMetricCounts(ctx, unitsKey, memberIDs, entries, metricUnits); err != nil {
			return nil, err
		}
	}

	generatedAt, err := s.getGeneratedAt(ctx, window, bucket)
	if err != nil {
		return nil, err
	}

	return &ports.GoodsLeaderboard{
		Board:       board,
		Window:      window,
		GeneratedAt: generatedAt,
		Entries:     entries,
	}, nil
}

func (s *Store) fillMetricCounts(
	ctx context.Context,
	key string,
	memberIDs []string,
	entries []ports.LeaderboardEntry,
	metric leaderboardMetric,
) error {
	values, err := s.client.Do(
		ctx,
		s.client.B().Zmscore().Key(key).Member(memberIDs...).Build(),
	).ToArray()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil
		}

		return fmt.Errorf("load leaderboard metric %s: %w", metric, err)
	}

	for i := range entries {
		if i >= len(values) {
			break
		}

		score, err := values[i].AsFloat64()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				continue
			}

			return fmt.Errorf("parse leaderboard metric %s: %w", metric, err)
		}

		switch metric {
		case metricOrders:
			entries[i].Orders = int64(score)
		case metricUnits:
			entries[i].Units = int64(score)
		}
	}

	return nil
}

func (s *Store) getGeneratedAt(ctx context.Context, window ports.LeaderboardWindow, bucket string) (*time.Time, error) {
	value, err := s.client.Do(
		ctx,
		s.client.B().Get().Key(generatedAtKey(window, bucket)).Build(),
	).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("get leaderboard generated_at: %w", err)
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, fmt.Errorf("parse leaderboard generated_at: %w", err)
	}

	return &parsed, nil
}

type aggregatedItem struct {
	goodID uuid.UUID
	gmv    decimal.Decimal
	orders int64
	units  int64
}

func aggregateOrderItems(items []ports.LeaderboardOrderItem) []aggregatedItem {
	type acc struct {
		gmv   decimal.Decimal
		units int64
	}

	byGood := make(map[uuid.UUID]acc, len(items))
	for _, item := range items {
		current := byGood[item.GoodID]
		current.gmv = current.gmv.Add(item.Price.Mul(decimal.NewFromInt32(item.Quantity)))
		current.units += int64(item.Quantity)
		byGood[item.GoodID] = current
	}

	result := make([]aggregatedItem, 0, len(byGood))
	for goodID, totals := range byGood {
		result = append(result, aggregatedItem{
			goodID: goodID,
			gmv:    totals.gmv,
			orders: 1,
			units:  totals.units,
		})
	}

	return result
}

type leaderboardMetric string

const (
	metricGMV    leaderboardMetric = "gmv"
	metricOrders leaderboardMetric = "orders"
	metricUnits  leaderboardMetric = "units"
)

func metricForBoard(board ports.LeaderboardBoard) (leaderboardMetric, error) {
	switch board {
	case ports.LeaderboardBoardGoodsGMV:
		return metricGMV, nil
	case ports.LeaderboardBoardGoodsOrders:
		return metricOrders, nil
	case ports.LeaderboardBoardGoodsUnits:
		return metricUnits, nil
	default:
		return "", ports.ErrUnsupportedLeaderboardBoard
	}
}

func scoreKey(metric leaderboardMetric, window ports.LeaderboardWindow, bucket string) string {
	return fmt.Sprintf("%s:%s:%s:%s", keyPrefix, metric, normalizeWindow(window), bucket)
}

func generatedAtKey(window ports.LeaderboardWindow, bucket string) string {
	return fmt.Sprintf("%s:generated_at:%s:%s", keyPrefix, normalizeWindow(window), bucket)
}

func processedEventKey(orderID uuid.UUID, aggregateVersion int32) string {
	return fmt.Sprintf("%s:processed:oms.order.completed.v1:%s:%d", keyPrefix, orderID.String(), aggregateVersion)
}

func bucketForWindow(window ports.LeaderboardWindow, now time.Time) string {
	switch window {
	case ports.LeaderboardWindowDay:
		return now.Format("2006-01-02")
	case ports.LeaderboardWindowWeek:
		year, week := now.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case ports.LeaderboardWindowMonth:
		return now.Format("2006-01")
	default:
		return now.Format("2006-01-02")
	}
}

func normalizeWindow(window ports.LeaderboardWindow) string {
	switch window {
	case ports.LeaderboardWindowDay:
		return "day"
	case ports.LeaderboardWindowWeek:
		return "week"
	case ports.LeaderboardWindowMonth:
		return "month"
	default:
		return "day"
	}
}
