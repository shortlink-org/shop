import { TAGS } from 'lib/constants';
import { describeFetchFailure, shopifyFetch } from '../fetch';
import { normalizeGood } from '../mappers';
import { GOODS_UNAVAILABLE } from '../sentinels';
import type {
  Good,
  ShopifyProductOperation,
  ShopifyProductRecommendationsOperation,
  ShopifyProductsOperation
} from '../types';
import { getGoodQuery, getGoodRecommendationsQuery, getGoodsQuery } from '../queries/good';

export type RequestOptions = { authorization?: string };

function filterGoodsByQuery(goods: Good[], query?: string): Good[] {
  const normalizedQuery = query?.trim().toLowerCase();
  if (!normalizedQuery) return goods;

  return goods.filter((good) => {
    const name = good.name.toLowerCase();
    const description = good.description.toLowerCase();
    return name.includes(normalizedQuery) || description.includes(normalizedQuery);
  });
}

function relevanceScore(good: Good, normalizedQuery: string): number {
  const name = good.name.toLowerCase();
  const description = good.description.toLowerCase();
  let score = 0;

  if (name === normalizedQuery) score += 5;
  if (name.startsWith(normalizedQuery)) score += 3;
  if (name.includes(normalizedQuery)) score += 2;
  if (description.includes(normalizedQuery)) score += 1;

  return score;
}

function sortGoods(
  goods: Good[],
  {
    sortKey,
    reverse,
    query
  }: {
    sortKey?: string;
    reverse?: boolean;
    query?: string;
  }
): Good[] {
  const sorted = [...goods];
  const normalizedQuery = query?.trim().toLowerCase() ?? '';

  if (sortKey === 'PRICE') {
    sorted.sort((a, b) => a.price - b.price);
  } else if (sortKey === 'CREATED_AT') {
    sorted.sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
  } else if (sortKey === 'RELEVANCE' && normalizedQuery) {
    sorted.sort((a, b) => relevanceScore(b, normalizedQuery) - relevanceScore(a, normalizedQuery));
  }

  if (reverse) {
    sorted.reverse();
  }

  return sorted;
}

export async function getGood(
  id: string,
  options?: RequestOptions
): Promise<Good | undefined | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyProductOperation>({
      query: getGoodQuery,
      variables: {
        id
      },
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    return res.body.data.good ? normalizeGood(res.body.data.good) : undefined;
  } catch (err) {
    console.error('[getGood] Failed to load good', { id, ...describeFetchFailure(err) });
    return GOODS_UNAVAILABLE;
  }
}

export async function getGoodRecommendations(
  id: string,
  options?: RequestOptions
): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyProductRecommendationsOperation>({
      query: getGoodRecommendationsQuery,
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    const recommendations = res.body.data.goods?.results ?? [];
    return recommendations.filter((good) => good.id !== id).map(normalizeGood);
  } catch (err) {
    console.error('[getGoodRecommendations] Failed to load recommendations', {
      id,
      ...describeFetchFailure(err)
    });
    return GOODS_UNAVAILABLE;
  }
}

export async function getGoods(
  {
    query,
    reverse,
    sortKey
  }: {
    query?: string;
    reverse?: boolean;
    sortKey?: string;
  },
  options?: RequestOptions
): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyProductsOperation>({
      cache: 'no-store',
      query: getGoodsQuery,
      tags: [TAGS.goods],
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    const goods = (res.body.data.goods?.results ?? []).map(normalizeGood);
    const filteredGoods = filterGoodsByQuery(goods, query);

    return sortGoods(filteredGoods, { sortKey, reverse, query });
  } catch (err) {
    console.error('[getGoods] Failed to load products', {
      query,
      sortKey,
      reverse,
      ...describeFetchFailure(err)
    });
    return GOODS_UNAVAILABLE;
  }
}
