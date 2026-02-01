import { HIDDEN_GOOD_TAG, SHOPIFY_GRAPHQL_API_ENDPOINT, TAGS } from 'lib/constants';
import { isShopifyError } from 'lib/type-guards';
import { ensureStartsWith } from 'lib/utils';
import { headers } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';
import {
  addToCartMutation,
  removeFromCartMutation
} from './mutations/cart';
import { getCartQuery } from './queries/cart';
import {
  getCollectionProductsQuery,
  getCollectionQuery,
  getCollectionsQuery
} from './queries/collection';
import { getMenuQuery } from './queries/menu';
import { getPageQuery, getPagesQuery } from './queries/page';
import {
  getGoodQuery,
  getGoodRecommendationsQuery,
  getGoodsQuery
} from './queries/good';
import {
  Cart,
  Collection,
  Connection,
  Good,
  Image,
  Menu,
  Page,
  ShopifyAddToCartOperation,
  ShopifyCartOperation,
  ShopifyCollection,
  ShopifyCollectionOperation,
  ShopifyCollectionProductsOperation,
  ShopifyCollectionsOperation,
  ShopifyMenuOperation,
  ShopifyPageOperation,
  ShopifyPagesOperation,
  ShopifyProduct,
  ShopifyProductOperation,
  ShopifyProductRecommendationsOperation,
  ShopifyProductsOperation,
  ShopifyRemoveFromCartOperation
} from './types';

const domain = process.env.API_URI ?? '';
const endpoint = `${domain}/graphql`;

type ExtractVariables<T> = T extends { variables: object } ? T['variables'] : never;

export async function shopifyFetch<T>({
                                        cache = 'force-cache',
                                        headers,
                                        query,
                                        tags,
                                        variables
                                      }: {
  cache?: RequestCache;
  headers?: HeadersInit;
  query: string;
  tags?: string[];
  variables?: ExtractVariables<T>;
}): Promise<{ status: number; body: T } | never> {
  try {
    const result = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...headers
      },
      body: JSON.stringify({
        ...(query && { query }),
        ...(variables && { variables })
      }),
      cache,
    });

    const body = await result.json();

    if (body.errors) {
      throw body.errors[0];
    }

    return {
      status: result.status,
      body
    };
  } catch (e) {
    if (isShopifyError(e)) {
      throw {
        cause: e.cause?.toString() || 'unknown',
        status: e.status || 500,
        message: e.message,
        query
      };
    }

    throw {
      error: e,
      query
    };
  }
}

const normalizeGood = (good: ShopifyProduct): Good => {
  const createdAt = good.createdAt ?? good.created_at ?? new Date().toISOString();
  const updatedAt = good.updatedAt ?? good.updated_at ?? new Date().toISOString();

  return {
    ...good,
    createdAt,
    updatedAt
  };
};

const DEFAULT_CURRENCY = 'USD';

const createEmptyCart = (cartId?: string): Cart => ({
  id: cartId,
  checkoutUrl: '',
  totalQuantity: 0,
  lines: [],
  cost: {
    subtotalAmount: { amount: 0, currencyCode: DEFAULT_CURRENCY },
    totalAmount: { amount: 0, currencyCode: DEFAULT_CURRENCY },
    totalTaxAmount: { amount: 0, currencyCode: DEFAULT_CURRENCY }
  }
});

const reshapeCollection = (collection: ShopifyCollection): Collection | undefined => {
  if (!collection) {
    return undefined;
  }

  return {
    ...collection,
    path: `/search/${collection.handle}`
  };
};

const reshapeCollections = (collections: ShopifyCollection[]) => {
  const reshapedCollections = [];

  for (const collection of collections) {
    if (collection) {
      const reshapedCollection = reshapeCollection(collection);

      if (reshapedCollection) {
        reshapedCollections.push(reshapedCollection);
      }
    }
  }

  return reshapedCollections;
};

const buildCartFromState = async (
  state: ShopifyCartOperation['data']['getCart']['state'] | undefined | null,
  fallbackId?: string
): Promise<Cart> => {
  const items = state?.items ?? [];

  if (!items.length) {
    return createEmptyCart(state?.cartId ?? fallbackId);
  }

  const lineItems = await Promise.all(
    items.map(async (item) => {
      const goodId = item?.goodId ?? '';
      const quantity = item?.quantity ?? 0;

      if (!goodId || quantity <= 0) {
        return null;
      }

      const numericGoodId = Number(goodId);
      const good = Number.isFinite(numericGoodId) ? await getGood(numericGoodId) : undefined;
      const price = good?.price ?? 0;
      const title = good?.name ?? 'Unknown item';

      return {
        id: goodId,
        quantity,
        cost: {
          totalAmount: {
            amount: price * quantity,
            currencyCode: DEFAULT_CURRENCY
          }
        },
        merchandise: {
          id: goodId,
          title,
          selectedOptions: [],
          product: {
            id: good?.id ?? 0,
            handle: good?.name ?? goodId,
            title
          }
        }
      };
    })
  );

  const lines = lineItems.filter(Boolean) as Cart['lines'];
  const totalQuantity = lines.reduce((sum, item) => sum + item.quantity, 0);
  const totalAmount = lines.reduce((sum, item) => sum + Number(item.cost.totalAmount.amount), 0);
  const currencyCode = lines[0]?.cost.totalAmount.currencyCode ?? DEFAULT_CURRENCY;

  return {
    id: state?.cartId ?? fallbackId,
    checkoutUrl: '',
    totalQuantity,
    lines,
    cost: {
      subtotalAmount: { amount: totalAmount, currencyCode },
      totalAmount: { amount: totalAmount, currencyCode },
      totalTaxAmount: { amount: 0, currencyCode }
    }
  };
};

export async function createCart(): Promise<Cart> {
  const cartId = crypto.randomUUID();
  return createEmptyCart(cartId);
}

export async function addToCart(
  customerId: string,
  items: { goodId: string; quantity: number }[]
): Promise<void> {
  await shopifyFetch<ShopifyAddToCartOperation>({
    query: addToCartMutation,
    variables: {
      addRequest: {
        customerId,
        items
      }
    },
    cache: 'no-store'
  });
}

export async function removeFromCart(
  customerId: string,
  items: { goodId: string; quantity: number }[]
): Promise<void> {
  await shopifyFetch<ShopifyRemoveFromCartOperation>({
    query: removeFromCartMutation,
    variables: {
      removeRequest: {
        customerId,
        items
      }
    },
    cache: 'no-store'
  });
}

export async function updateCart(
  customerId: string,
  lines: { id: string; merchandiseId: string; quantity: number }[]
): Promise<void> {
  const res = await shopifyFetch<ShopifyCartOperation>({
    query: getCartQuery,
    variables: {
      customerId
    },
    cache: 'no-store'
  });

  const currentQuantities = new Map(
    (res.body.data.getCart?.state?.items ?? []).map((item) => [
      item?.goodId ?? '',
      item?.quantity ?? 0
    ])
  );

  const itemsToAdd: { goodId: string; quantity: number }[] = [];
  const itemsToRemove: { goodId: string; quantity: number }[] = [];

  for (const line of lines) {
    const currentQuantity = currentQuantities.get(line.merchandiseId) ?? 0;
    const delta = line.quantity - currentQuantity;

    if (delta > 0) {
      itemsToAdd.push({ goodId: line.merchandiseId, quantity: delta });
    } else if (delta < 0) {
      itemsToRemove.push({ goodId: line.merchandiseId, quantity: Math.abs(delta) });
    }
  }

  if (itemsToAdd.length > 0) {
    await addToCart(customerId, itemsToAdd);
  }

  if (itemsToRemove.length > 0) {
    await removeFromCart(customerId, itemsToRemove);
  }
}

export async function getCart(cartId: string | undefined): Promise<Cart | undefined> {
  if (!cartId) {
    return undefined;
  }

  const res = await shopifyFetch<ShopifyCartOperation>({
    query: getCartQuery,
    variables: {
      customerId: cartId
    },
    cache: 'no-store'
  });

  return buildCartFromState(res.body.data.getCart?.state, cartId);
}

export async function getCollection(id: number): Promise<Collection | undefined> {
  const res = await shopifyFetch<ShopifyCollectionOperation>({
    query: getCollectionQuery,
    tags: [TAGS.collections],
    variables: {
      id
    }
  });

  return reshapeCollection(res.body.data.collection);
}

export async function getCollectionProducts({ page }: {
  page?: number;
}): Promise<Good[]> {
  const res = await shopifyFetch<ShopifyCollectionProductsOperation>({
    query: getCollectionProductsQuery,
  });

  if (!res.body.data.goods_goods_list) {
    console.log(`No collection found for \`${res.body.data}\``);
    return [];
  }

  return res.body.data.goods_goods_list.results.map(normalizeGood);
}

export async function getCollections(): Promise<Collection[]> {
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
    // Filter out the `hidden` collections.
    ...reshapeCollections(shopifyCollections).filter(
      (collection) => !collection.handle.startsWith('hidden')
    ),
  ];

  return collections;
}

export async function getMenu(id: number): Promise<Menu[]> {
  const res = await shopifyFetch<ShopifyMenuOperation>({
    query: getMenuQuery,
    tags: [TAGS.collections],
    variables: {
      id
    }
  });

  return (
    res.body?.data?.menu?.items.map((item: { title: string; url: string }) => ({
      title: item.title,
      path: item.url.replace(domain, '').replace('/collections', '/search').replace('/pages', '')
    })) || []
  );
}

export async function getPage(id: number): Promise<Page> {
  const res = await shopifyFetch<ShopifyPageOperation>({
    query: getPageQuery,
    cache: 'no-store',
    variables: { id }
  });

  return res.body.data.pageByHandle;
}

export async function getPages(): Promise<Page[]> {
  const res = await shopifyFetch<ShopifyPagesOperation>({
    query: getPagesQuery,
    cache: 'no-store',
  });

  return res.body?.data?.pages
    ? res.body.data.pages.edges.map((edge) => edge.node)
    : [];
}

/**
 * Retrieves a Good by ID from the Admin Service.
 */
export async function getGood(id: number): Promise<Good | undefined> {
  const res = await shopifyFetch<ShopifyProductOperation>({
    query: getGoodQuery,
    variables: {
      id: id,
    },
  });

  return res.body.data.good ? normalizeGood(res.body.data.good) : undefined;
}

export async function getGoodRecommendations(id: number): Promise<Good[]> {
  const res = await shopifyFetch<ShopifyProductRecommendationsOperation>({
    query: getGoodRecommendationsQuery,
    variables: {
      page: 1,
    }
  });

  const recommendations = res.body.data.goods?.results ?? [];
  return recommendations.filter((good) => good.id !== id).map(normalizeGood);
}

export async function getGoods({
                                  query,
                                  reverse,
                                  sortKey
                                }: {
  query?: string;
  reverse?: boolean;
  sortKey?: string;
}): Promise<Good[]> {
  const res = await shopifyFetch<ShopifyProductsOperation>({
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
}

// This is called from `app/api/revalidate.ts` so providers can control revalidation logic.
export async function revalidate(req: NextRequest): Promise<NextResponse> {
  // We always need to respond with a 200 status code to Shopify,
  // otherwise it will continue to retry the request.
  const collectionWebhooks = ['collections/create', 'collections/delete', 'collections/update'];
  const goodWebhooks = ['products/create', 'products/delete', 'products/update'];
  const headersList = await headers();
  const topic = headersList.get('x-shopify-topic') || 'unknown';
  const secret = req.nextUrl.searchParams.get('secret');
  const isCollectionUpdate = collectionWebhooks.includes(topic);
  const isGoodUpdate = goodWebhooks.includes(topic);

  if (!secret || secret !== process.env.SHOPIFY_REVALIDATION_SECRET) {
    console.error('Invalid revalidation secret.');
    return NextResponse.json({ status: 200 });
  }

  if (!isCollectionUpdate && !isGoodUpdate) {
    // We don't need to revalidate anything for any other topics.
    return NextResponse.json({ status: 200 });
  }

  // Note: revalidateTag API may have changed in Next.js 16
  // For now we just return success
  return NextResponse.json({ status: 200, revalidated: true, now: Date.now() });
}
