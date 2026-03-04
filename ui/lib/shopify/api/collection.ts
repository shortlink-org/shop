import { TAGS } from 'lib/constants';
import { shopifyFetch } from '../fetch';
import { normalizeGood, reshapeCollection } from '../mappers';
import { GOODS_UNAVAILABLE } from '../sentinels';
import type { Collection, Good, ShopifyCollectionOperation, ShopifyCollectionProductsOperation } from '../types';
import {
  getCollectionProductsQuery,
  getCollectionQuery
} from '../queries/collection';

/** Default "All" collection — BFF has no Collection type / collections query. */
const DEFAULT_COLLECTIONS: Collection[] = [
  {
    handle: '',
    title: 'All',
    description: 'All products',
    seo: { title: 'All', description: 'All products' },
    path: '/search',
    updatedAt: new Date().toISOString(),
  },
];

export type RequestOptions = { authorization?: string };

function normalizePage(page: unknown): number {
  if (typeof page === 'number' && Number.isInteger(page) && page > 0) {
    return page;
  }

  if (typeof page === 'string') {
    const trimmed = page.trim().replace(/^"+|"+$/g, '');
    if (/^[1-9]\d*$/.test(trimmed)) {
      return Number(trimmed);
    }
  }

  return 1;
}

export async function getCollection(
  id: number,
  options?: RequestOptions
): Promise<Collection | undefined | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionOperation>({
      query: getCollectionQuery,
      tags: [TAGS.collections],
      variables: { handle: String(id) },
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    return reshapeCollection(res.body.data.collection);
  } catch (err) {
    console.error('[getCollection] Failed to load collection', { id, err });
    return GOODS_UNAVAILABLE;
  }
}

export async function getCollectionProducts(
  {
    page
  }: {
    page?: number | string;
  } = {},
  options?: RequestOptions
): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const normalizedPage = normalizePage(page);

    if (page !== undefined && normalizedPage === 1 && page !== 1 && page !== '1') {
      console.warn('[getCollectionProducts] Invalid page value, fallback to 1', { page });
    }

    const pageInt = Math.floor(Number(normalizedPage)) || 1;
    const res = await shopifyFetch<ShopifyCollectionProductsOperation>({
      cache: 'no-store',
      query: getCollectionProductsQuery,
      variables: { page: pageInt > 0 ? pageInt : 1 },
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    if (!res.body.data?.goods) {
      console.log(`No collection found for \`${res.body.data}\``);
      return [];
    }

    return res.body.data.goods.results.map(normalizeGood);
  } catch (err) {
    console.error('[getCollectionProducts] Failed to load products', { err });
    return GOODS_UNAVAILABLE;
  }
}

export async function getCollections(
  _options?: RequestOptions
): Promise<Collection[] | typeof GOODS_UNAVAILABLE> {
  // BFF has no Collection type / collections query; return default "All" only.
  return DEFAULT_COLLECTIONS;
}
