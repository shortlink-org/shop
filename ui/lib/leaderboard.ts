import type {
  LeaderboardEntry,
  LeaderboardFilter,
  LeaderboardFilterId,
  LeaderboardStat
} from '@shortlink-org/ui-kit';

function avatar(imageId: number): string {
  return `https://i.pravatar.cc/240?img=${imageId}`;
}

export const leaderboardFilters: LeaderboardFilter[] = [
  { id: 'day', label: 'Today' },
  { id: 'week', label: 'This week' },
  { id: 'month', label: 'This month' }
];

export const leaderboardEntries: Record<LeaderboardFilterId, LeaderboardEntry[]> = {
  day: [
    {
      id: 'northstar',
      rank: 1,
      name: 'Northstar Atelier',
      subtitle: 'Premium home decor',
      avatarSrc: avatar(12),
      score: 128400,
      delta: 2,
      metric: '1.2k orders',
      note: 'Same-day delivery leader',
      verified: true,
      badge: { label: 'Top conversion', tone: 'accent' },
      accentColor: 'rgba(56, 189, 248, 0.28)'
    },
    {
      id: 'luma',
      rank: 2,
      name: 'Luma Skin Lab',
      subtitle: 'Beauty marketplace standout',
      avatarSrc: avatar(32),
      score: 121900,
      delta: 1,
      metric: '8.6% CVR',
      note: 'Best paid media ROAS',
      verified: true,
      badge: { label: 'Fast mover', tone: 'success' },
      accentColor: 'rgba(244, 114, 182, 0.24)'
    },
    {
      id: 'orbit',
      rank: 3,
      name: 'Orbit Run Club',
      subtitle: 'Sneakers and active gear',
      avatarSrc: avatar(20),
      score: 119300,
      delta: -1,
      metric: '934 units',
      note: 'Top basket size',
      badge: { label: 'Bundle king', tone: 'warning' },
      accentColor: 'rgba(52, 211, 153, 0.24)'
    },
    {
      id: 'atelier-noir',
      rank: 4,
      name: 'Atelier Noir',
      subtitle: 'Designer accessories',
      avatarSrc: avatar(48),
      score: 107200,
      delta: 3,
      metric: '17 live products',
      note: 'Campaign spike',
      verified: true
    },
    {
      id: 'fields',
      rank: 5,
      name: 'Fields Pantry',
      subtitle: 'Gourmet grocery',
      avatarSrc: avatar(14),
      score: 101300,
      delta: 1,
      metric: '312 repeat buyers',
      note: 'Highest repeat rate'
    },
    {
      id: 'aurora',
      rank: 6,
      name: 'Aurora Frames',
      subtitle: 'Art and prints',
      avatarSrc: avatar(38),
      score: 94800,
      delta: -2,
      metric: '42% gross margin',
      note: 'Premium assortment'
    },
    {
      id: 'sora',
      rank: 7,
      name: 'Sora Tea Studio',
      subtitle: 'Tea and ritual sets',
      avatarSrc: avatar(45),
      score: 84700,
      delta: 2,
      metric: '4.9 merchant rating',
      note: 'Review momentum'
    },
    {
      id: 'user',
      rank: 11,
      name: 'Your store',
      subtitle: 'Curated stationery',
      avatarSrc: avatar(55),
      score: 56300,
      delta: 4,
      metric: '18.2% growth',
      note: 'Current campaign',
      isCurrentUser: true,
      badge: { label: 'On the rise', tone: 'success' }
    }
  ],
  week: [
    {
      id: 'luma',
      rank: 1,
      name: 'Luma Skin Lab',
      subtitle: 'Beauty marketplace standout',
      avatarSrc: avatar(32),
      score: 824200,
      delta: 1,
      metric: '11.3k orders',
      note: 'Highest repeat purchase',
      verified: true,
      badge: { label: 'Best retention', tone: 'accent' },
      accentColor: 'rgba(244, 114, 182, 0.24)'
    },
    {
      id: 'northstar',
      rank: 2,
      name: 'Northstar Atelier',
      subtitle: 'Premium home decor',
      avatarSrc: avatar(12),
      score: 801100,
      delta: -1,
      metric: '9.8k orders',
      note: 'Strong cross-sell',
      verified: true,
      badge: { label: 'Big baskets', tone: 'warning' },
      accentColor: 'rgba(56, 189, 248, 0.28)'
    },
    {
      id: 'orbit',
      rank: 3,
      name: 'Orbit Run Club',
      subtitle: 'Sneakers and active gear',
      avatarSrc: avatar(20),
      score: 768000,
      delta: 2,
      metric: '6.4k units',
      note: 'Weekend spike',
      badge: { label: 'Momentum', tone: 'success' },
      accentColor: 'rgba(52, 211, 153, 0.24)'
    },
    {
      id: 'cinder',
      rank: 4,
      name: 'Cinder Supply',
      subtitle: 'Industrial style furniture',
      avatarSrc: avatar(5),
      score: 645400,
      delta: 1,
      metric: '92 AOV',
      note: 'High-ticket growth'
    },
    {
      id: 'atelier-noir',
      rank: 5,
      name: 'Atelier Noir',
      subtitle: 'Designer accessories',
      avatarSrc: avatar(48),
      score: 602300,
      delta: -2,
      metric: '28% promo mix',
      note: 'Luxury segment',
      verified: true
    },
    {
      id: 'fields',
      rank: 6,
      name: 'Fields Pantry',
      subtitle: 'Gourmet grocery',
      avatarSrc: avatar(14),
      score: 589800,
      delta: 3,
      metric: '2.7k subscriptions',
      note: 'Subscription boost'
    },
    {
      id: 'sora',
      rank: 7,
      name: 'Sora Tea Studio',
      subtitle: 'Tea and ritual sets',
      avatarSrc: avatar(45),
      score: 521900,
      delta: 1,
      metric: '4.95 merchant rating',
      note: 'Community favorite'
    },
    {
      id: 'user',
      rank: 10,
      name: 'Your store',
      subtitle: 'Curated stationery',
      avatarSrc: avatar(55),
      score: 348900,
      delta: 2,
      metric: '7.1% CVR',
      note: 'Current campaign',
      isCurrentUser: true,
      badge: { label: 'Closing in', tone: 'success' }
    }
  ],
  month: [
    {
      id: 'northstar',
      rank: 1,
      name: 'Northstar Atelier',
      subtitle: 'Premium home decor',
      avatarSrc: avatar(12),
      score: 3210000,
      delta: 0,
      metric: '28.4k orders',
      note: 'Marketplace staple',
      verified: true,
      badge: { label: 'Marketplace icon', tone: 'accent' },
      accentColor: 'rgba(56, 189, 248, 0.28)'
    },
    {
      id: 'luma',
      rank: 2,
      name: 'Luma Skin Lab',
      subtitle: 'Beauty marketplace standout',
      avatarSrc: avatar(32),
      score: 3040000,
      delta: 0,
      metric: '23.9k orders',
      note: 'Retention leader',
      verified: true,
      badge: { label: 'Fan favorite', tone: 'success' },
      accentColor: 'rgba(244, 114, 182, 0.24)'
    },
    {
      id: 'orbit',
      rank: 3,
      name: 'Orbit Run Club',
      subtitle: 'Sneakers and active gear',
      avatarSrc: avatar(20),
      score: 2860000,
      delta: 1,
      metric: '18.1k units',
      note: 'Strong launch cycle',
      badge: { label: 'Seasonal peak', tone: 'warning' },
      accentColor: 'rgba(52, 211, 153, 0.24)'
    },
    {
      id: 'atelier-noir',
      rank: 4,
      name: 'Atelier Noir',
      subtitle: 'Designer accessories',
      avatarSrc: avatar(48),
      score: 2400000,
      delta: -1,
      metric: '52% margin',
      note: 'Luxury audience',
      verified: true
    },
    {
      id: 'cinder',
      rank: 5,
      name: 'Cinder Supply',
      subtitle: 'Industrial style furniture',
      avatarSrc: avatar(5),
      score: 2180000,
      delta: 2,
      metric: '$182 AOV',
      note: 'High-ticket gain'
    },
    {
      id: 'fields',
      rank: 6,
      name: 'Fields Pantry',
      subtitle: 'Gourmet grocery',
      avatarSrc: avatar(14),
      score: 2070000,
      delta: -1,
      metric: '5.2k subscribers',
      note: 'Subscription engine'
    },
    {
      id: 'aurora',
      rank: 7,
      name: 'Aurora Frames',
      subtitle: 'Art and prints',
      avatarSrc: avatar(38),
      score: 1940000,
      delta: 1,
      metric: '31% returning buyers',
      note: 'Collector audience'
    },
    {
      id: 'user',
      rank: 9,
      name: 'Your store',
      subtitle: 'Curated stationery',
      avatarSrc: avatar(55),
      score: 1620000,
      delta: 3,
      metric: '24% monthly growth',
      note: 'Current campaign',
      isCurrentUser: true,
      badge: { label: 'Breakout month', tone: 'success' }
    }
  ]
};

export const leaderboardStats: Record<LeaderboardFilterId, LeaderboardStat[]> = {
  day: [
    { id: 'gmv', label: 'Tracked GMV', value: 1920000, change: '+18%', tone: 'accent' },
    { id: 'buyers', label: 'Active buyers', value: 14820, change: '+9%', tone: 'success' },
    {
      id: 'conversion',
      label: 'Avg. conversion',
      value: '7.4%',
      change: '+0.6pp',
      tone: 'warning'
    },
    { id: 'campaigns', label: 'Campaigns live', value: 27, change: 'today', tone: 'neutral' }
  ],
  week: [
    { id: 'gmv', label: 'Tracked GMV', value: 12400000, change: '+26%', tone: 'accent' },
    { id: 'buyers', label: 'Active buyers', value: 58240, change: '+13%', tone: 'success' },
    {
      id: 'conversion',
      label: 'Avg. conversion',
      value: '6.8%',
      change: '+0.4pp',
      tone: 'warning'
    },
    { id: 'campaigns', label: 'Campaigns live', value: 64, change: 'this week', tone: 'neutral' }
  ],
  month: [
    { id: 'gmv', label: 'Tracked GMV', value: 48600000, change: '+41%', tone: 'accent' },
    { id: 'buyers', label: 'Active buyers', value: 201400, change: '+22%', tone: 'success' },
    {
      id: 'conversion',
      label: 'Avg. conversion',
      value: '6.4%',
      change: '+0.3pp',
      tone: 'warning'
    },
    { id: 'campaigns', label: 'Campaigns live', value: 112, change: 'this month', tone: 'neutral' }
  ]
};
