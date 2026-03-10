'use client';

import { Button, MarketplaceLeaderboard } from '@shortlink-org/ui-kit';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { leaderboardEntries, leaderboardFilters, leaderboardStats } from 'lib/leaderboard';

export function ShopLeaderboard() {
  const [selectedFilterId, setSelectedFilterId] = useState<keyof typeof leaderboardEntries>('week');

  const entries = useMemo(
    () => leaderboardEntries[selectedFilterId] || leaderboardEntries.week || [],
    [selectedFilterId]
  );
  const stats = useMemo(
    () => leaderboardStats[selectedFilterId] || leaderboardStats.week || [],
    [selectedFilterId]
  );

  return (
    <MarketplaceLeaderboard
      eyebrow="Marketplace leaderboard"
      title="Who is setting the pace this week"
      description="A live view into the storefronts leading on conversion, repeat buyers and revenue momentum, now framed as part of the storefront instead of a separate dashboard."
      scoreLabel="GMV"
      entries={entries}
      stats={stats}
      filters={leaderboardFilters}
      selectedFilterId={selectedFilterId}
      onFilterChange={setSelectedFilterId}
      visibleRows={7}
      headerAction={
        <Button as={Link} asProps={{ href: '/search?sort=trending' }} variant="secondary">
          Browse trending goods
        </Button>
      }
      className="rounded-[2rem] border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_92%,transparent)] shadow-[0_30px_80px_-52px_rgba(15,23,42,0.38)]"
    />
  );
}
