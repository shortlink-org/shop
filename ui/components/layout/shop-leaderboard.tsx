'use client';

import { Button, MarketplaceLeaderboard } from '@shortlink-org/ui-kit';
import Link from 'next/link';
import { useEffect, useReducer, useState } from 'react';
import type { LeaderboardEntry, LeaderboardStat } from '@shortlink-org/ui-kit';
import { leaderboardFilters, loadGoodsLeaderboard } from 'lib/leaderboard';

type LeaderboardState = {
  entries: LeaderboardEntry[];
  stats: LeaderboardStat[];
  loading: boolean;
};

function leaderboardReducer(
  state: LeaderboardState,
  action:
    | { type: 'LOADING' }
    | { type: 'SUCCESS'; payload: { entries: LeaderboardEntry[]; stats: LeaderboardStat[] } }
    | { type: 'ERROR' }
): LeaderboardState {
  switch (action.type) {
    case 'LOADING':
      return { ...state, loading: true };
    case 'SUCCESS':
      return { entries: action.payload.entries, stats: action.payload.stats, loading: false };
    case 'ERROR':
      return { entries: [], stats: [], loading: false };
    default:
      return state;
  }
}

export function ShopLeaderboard() {
  const [selectedFilterId, setSelectedFilterId] = useState<'day' | 'week' | 'month'>('week');
  const [state, dispatch] = useReducer(leaderboardReducer, {
    entries: [],
    stats: [],
    loading: true
  });

  useEffect(() => {
    let cancelled = false;
    dispatch({ type: 'LOADING' });

    loadGoodsLeaderboard(selectedFilterId)
      .then((payload) => {
        if (cancelled) return;
        dispatch({ type: 'SUCCESS', payload: { entries: payload.entries, stats: payload.stats } });
      })
      .catch(() => {
        if (cancelled) return;
        dispatch({ type: 'ERROR' });
      });

    return () => {
      cancelled = true;
    };
  }, [selectedFilterId]);

  function handleFilterChange(filterId: string) {
    if (filterId === 'day' || filterId === 'week' || filterId === 'month') {
      setSelectedFilterId(filterId);
    }
  }

  return (
    <MarketplaceLeaderboard
      eyebrow="Goods leaderboard"
      title="Which goods are setting the pace"
      description="A live read model built from completed orders, ranking the items driving the most revenue over the current period."
      scoreLabel="GMV"
      entries={state.entries}
      stats={state.stats}
      filters={leaderboardFilters}
      selectedFilterId={selectedFilterId}
      onFilterChange={handleFilterChange}
      visibleRows={7}
      loading={state.loading}
      emptyTitle="No completed orders yet"
      emptyDescription="The leaderboard will populate once the first completed purchases land in OMS."
      headerAction={
        <Button as={Link} asProps={{ href: '/search?sort=trending' }} variant="secondary">
          Browse trending goods
        </Button>
      }
      className="rounded-[2rem] border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_92%,transparent)] shadow-[0_30px_80px_-52px_rgba(15,23,42,0.38)]"
    />
  );
}
