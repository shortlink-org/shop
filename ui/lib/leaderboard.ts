import type {
  LeaderboardEntry,
  LeaderboardFilter,
  LeaderboardFilterId,
  LeaderboardStat
} from '@shortlink-org/ui-kit';
import { getGoodsLeaderboard, getGood, GOODS_UNAVAILABLE } from 'lib/shopify';

export const leaderboardFilters: LeaderboardFilter[] = [
  { id: 'day', label: 'Today' },
  { id: 'week', label: 'This week' },
  { id: 'month', label: 'This month' }
];

type LeaderboardPayload = {
  entries: LeaderboardEntry[];
  stats: LeaderboardStat[];
};

export async function loadGoodsLeaderboard(
  filterId: LeaderboardFilterId
): Promise<LeaderboardPayload> {
  const window = filterIdToWindow(filterId);
  const leaderboard = await getGoodsLeaderboard(window, 7);

  if (!leaderboard?.entries?.length) {
    return {
      entries: [],
      stats: []
    };
  }

  const goods = await Promise.all(
    leaderboard.entries.map(async (entry) => {
      const goodId = entry?.memberId ?? '';
      if (!goodId) {
        return null;
      }

      const good = await getGood(goodId);
      return good === GOODS_UNAVAILABLE ? null : good;
    })
  );

  const mappedEntries = leaderboard.entries
    .map((entry, index) => {
      const goodId = entry?.memberId ?? '';
      if (!goodId) {
        return null;
      }

      const good = goods[index];
      const score = entry?.score ?? 0;
      const orders = entry?.orders ?? 0;
      const units = entry?.units ?? 0;

      return {
        id: goodId,
        rank: entry?.rank ?? index + 1,
        name: good?.name ?? `Good ${goodId.slice(0, 8)}`,
        subtitle: `${orders} orders · ${units} units`,
        href: `/good/${goodId}`,
        score,
        scoreDisplay: formatCurrency(score),
        metric: `${units} units`,
        note: good?.description?.slice(0, 72) || 'Top-performing item this period'
      } satisfies LeaderboardEntry;
    })
    .filter((entry) => entry !== null);

  const entries: LeaderboardEntry[] = mappedEntries;

  return {
    entries,
    stats: buildStats(entries)
  };
}

function buildStats(entries: LeaderboardEntry[]): LeaderboardStat[] {
  if (entries.length === 0) {
    return [];
  }

  const totalScore = entries.reduce((sum, entry) => sum + entry.score, 0);
  const topScore = entries[0]?.score ?? 0;

  return [
    {
      id: 'top-score',
      label: 'Top GMV',
      value: formatCurrency(topScore),
      tone: 'accent'
    },
    {
      id: 'tracked-goods',
      label: 'Visible goods',
      value: entries.length,
      tone: 'neutral'
    },
    {
      id: 'sum-score',
      label: 'Visible GMV',
      value: formatCurrency(totalScore),
      tone: 'success'
    }
  ];
}

function filterIdToWindow(filterId: LeaderboardFilterId): 'DAY' | 'WEEK' | 'MONTH' {
  switch (filterId) {
    case 'day':
      return 'DAY';
    case 'month':
      return 'MONTH';
    default:
      return 'WEEK';
  }
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0
  }).format(value);
}
