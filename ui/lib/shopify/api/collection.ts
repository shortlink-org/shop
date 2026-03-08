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
  options?: RequestOptions
): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  const headers: HeadersInit | undefined = options?.authorization
    ? { Authorization: options.authorization }
    : undefined;
  try {
    const res = await shopifyFetch<ShopifyCollectionProductsOperation>({
      cache: 'no-store',
      query: getCollectionProductsQuery,
      headers
    });

    if (!res.body.data?.goods) {
      console.log(`No collection found for \`${res.body.data}\``);
      return [];
    }

    return res.body.data.goods.results.map(normalizeGood);
  } catch (err) {
    const errorPath = extractErrorPath(err);
    const traceId = extractTraceId(err);
    console.error('[getCollectionProducts] Failed to load products', { err, errorPath, traceId });
    return GOODS_UNAVAILABLE;
  }
}

export async function getCollections(
  _options?: RequestOptions
): Promise<Collection[] | typeof GOODS_UNAVAILABLE> {
  // BFF has no Collection type / collections query; return default "All" only.
  return DEFAULT_COLLECTIONS;
}

function extractErrorPath(err: unknown): string | undefined {
  const path = getNestedProperty(err, 'path');
  return path !== undefined ? JSON.stringify(path) : undefined;
}

function extractTraceId(err: unknown): string | undefined {
  const traceId = getNestedProperty(err, 'traceId');
  if (typeof traceId === 'string' && traceId.trim()) {
    return traceId;
  }

  const extensions = getNestedProperty(err, 'extensions');
  if (typeof extensions === 'object' && extensions !== null) {
    const extensionTraceId =
      ('traceId' in extensions && typeof extensions.traceId === 'string' && extensions.traceId) ||
      ('traceID' in extensions && typeof extensions.traceID === 'string' && extensions.traceID) ||
      ('trace_id' in extensions && typeof extensions.trace_id === 'string' && extensions.trace_id);
    return extensionTraceId || undefined;
  }

  return undefined;
}

function getNestedProperty(value: unknown, key: string): unknown {
  if (typeof value !== 'object' || value === null) return undefined;
  if (key in value) return (value as Record<string, unknown>)[key];
  if ('error' in value) return getNestedProperty((value as { error?: unknown }).error, key);
  return undefined;
}
