'use client';

import { Button, MarketplaceLeaderboard } from '@shortlink-org/ui-kit';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { leaderboardEntries, leaderboardFilters, leaderboardStats } from 'lib/leaderboard';

export function ShopLeaderboard() {
  const [selectedFilterId, setSelectedFilterId] =
    useState<keyof typeof leaderboardEntries>('week');

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
      title="Top shops this season"
      description="A live look at the storefronts setting the pace across conversion, repeat buyers and revenue momentum."
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
      className="rounded-[2rem]"
    />
  );
}
