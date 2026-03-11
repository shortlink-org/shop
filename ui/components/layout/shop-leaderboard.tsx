'use client';

import { Button, MarketplaceLeaderboard } from '@shortlink-org/ui-kit';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import type { LeaderboardEntry, LeaderboardStat } from '@shortlink-org/ui-kit';
import { leaderboardFilters, loadGoodsLeaderboard } from 'lib/leaderboard';

export function ShopLeaderboard() {
  const [selectedFilterId, setSelectedFilterId] = useState<'day' | 'week' | 'month'>('week');
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [stats, setStats] = useState<LeaderboardStat[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    setLoading(true);
    loadGoodsLeaderboard(selectedFilterId)
      .then((payload) => {
        if (cancelled) {
          return;
        }

        setEntries(payload.entries);
        setStats(payload.stats);
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setEntries([]);
        setStats([]);
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
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
      entries={entries}
      stats={stats}
      filters={leaderboardFilters}
      selectedFilterId={selectedFilterId}
      onFilterChange={handleFilterChange}
      visibleRows={7}
      loading={loading}
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
