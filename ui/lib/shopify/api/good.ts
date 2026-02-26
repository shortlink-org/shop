import { TAGS } from 'lib/constants';
import { shopifyFetch } from '../fetch';
import { normalizeGood } from '../mappers';
import { GOODS_UNAVAILABLE } from '../sentinels';
import type { Good, ShopifyProductOperation, ShopifyProductRecommendationsOperation, ShopifyProductsOperation } from '../types';
import { getGoodQuery, getGoodRecommendationsQuery, getGoodsQuery } from '../queries/good';

export async function getGood(
  id: number
): Promise<Good | undefined | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyProductOperation>({
      query: getGoodQuery,
      variables: {
        id: id,
      },
    });

    return res.body.data.good ? normalizeGood(res.body.data.good) : undefined;
  } catch (err) {
    console.error('[getGood] Failed to load good', { id, err });
    return GOODS_UNAVAILABLE;
  }
}

export async function getGoodRecommendations(
  id: number
): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyProductRecommendationsOperation>({
      query: getGoodRecommendationsQuery,
      variables: {
        page: 1,
      }
    });

    const recommendations = res.body.data.goods?.results ?? [];
    return recommendations.filter((good) => good.id !== id).map(normalizeGood);
  } catch (err) {
    console.error('[getGoodRecommendations] Failed to load recommendations', { id, err });
    return GOODS_UNAVAILABLE;
  }
}

export async function getGoods({
  query,
  reverse,
  sortKey
}: {
  query?: string;
  reverse?: boolean;
  sortKey?: string;
}): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyProductsOperation>({
      cache: 'no-store',
      query: getGoodsQuery,
      tags: [TAGS.goods],
      variables: {
        query,
        reverse,
        sortKey
      }
    });

    const goods = res.body.data.goods?.results ?? [];
    return goods.map(normalizeGood);
  } catch (err) {
    console.error('[getGoods] Failed to load products', { query, sortKey, reverse, err });
    return GOODS_UNAVAILABLE;
  }
}
