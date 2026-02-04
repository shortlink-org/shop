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

export async function getCollection(
  id: number
): Promise<Collection | undefined | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionOperation>({
      query: getCollectionQuery,
      tags: [TAGS.collections],
      variables: {
        id
      }
    });

    return reshapeCollection(res.body.data.collection);
  } catch (err) {
    console.error('[getCollection] Failed to load collection', { id, err });
    return GOODS_UNAVAILABLE;
  }
}

export async function getCollectionProducts({
  page
}: {
  page?: number;
}): Promise<Good[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionProductsOperation>({
      query: getCollectionProductsQuery,
    });

    if (!res.body.data.goods_goods_list) {
      console.log(`No collection found for \`${res.body.data}\``);
      return [];
    }

    return res.body.data.goods_goods_list.results.map(normalizeGood);
  } catch (err) {
    console.error('[getCollectionProducts] Failed to load products', { err });
    return GOODS_UNAVAILABLE;
  }
}

export async function getCollections(): Promise<Collection[] | typeof GOODS_UNAVAILABLE> {
  try {
    const res = await shopifyFetch<ShopifyCollectionsOperation>({
      query: getCollectionsQuery,
      tags: [TAGS.collections],
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
