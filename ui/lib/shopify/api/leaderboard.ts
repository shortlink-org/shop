import { shopifyFetch } from '../fetch';
import type { ShopifyLeaderboardOperation } from '../types';
import { getLeaderboardQuery } from '../queries/leaderboard';

export async function getGoodsLeaderboard(window: 'DAY' | 'WEEK' | 'MONTH', limit = 7) {
  const res = await shopifyFetch<ShopifyLeaderboardOperation>({
    cache: 'no-store',
    query: getLeaderboardQuery,
    variables: {
      board: 'GOODS_GMV',
      window,
      limit
    }
  });

  return res.body.data.getLeaderboard?.leaderboard ?? null;
}
