import { TAGS } from 'lib/constants';
import { shopifyFetch } from '../fetch';
import { normalizeGood, reshapeCollection, reshapeCollections } from '../mappers';
import { GOODS_UNAVAILABLE } from '../sentinels';
import type { Collection, Good, ShopifyCollectionOperation, ShopifyCollectionProductsOperation, ShopifyCollectionsOperation } from '../types';
import {
  getCollectionProductsQuery,
  getCollectionQuery,
  getCollectionsQuery
} from '../queries/collection';

export type RequestOptions = { authorization?: string };

export async function getCollection(
  id: number,
  options?: RequestOptions
): Promise<Collection | undefined | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionOperation>({
      query: getCollectionQuery,
      tags: [TAGS.collections],
      variables: {
        id
      },
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
    page?: number;
  } = {},
  options?: RequestOptions
): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionProductsOperation>({
      cache: 'no-store',
      query: getCollectionProductsQuery,
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
  options?: RequestOptions
): Promise<Collection[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionsOperation>({
      query: getCollectionsQuery,
      tags: [TAGS.collections],
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    const shopifyCollections = res.body?.data?.collections
      ? res.body.data.collections.edges.map((edge) => edge.node)
      : [];

    const collections = [
      {
        handle: '',
        title: 'All',
        description: 'All products',
        seo: {
          title: 'All',
          description: 'All products',
        },
        path: '/search',
        updatedAt: new Date().toISOString(),
      },
      ...reshapeCollections(shopifyCollections).filter(
        (collection) => !collection.handle.startsWith('hidden')
      ),
    ];

    return collections;
  } catch (err) {
    console.error('[getCollections] Failed to load collections', { err });
    return GOODS_UNAVAILABLE;
  }
}
