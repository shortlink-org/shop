import { HIDDEN_GOOD_TAG, SHOPIFY_GRAPHQL_API_ENDPOINT, TAGS } from 'lib/constants';
import { isShopifyError } from 'lib/type-guards';
import { ensureStartsWith } from 'lib/utils';
import {
  addToCartMutation,
  removeFromCartMutation,
  checkoutMutation
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
  CheckoutInput,
  CheckoutResult,
  Collection,
  Connection,
  Good,
  Image,
  Menu,
  Page,
  ShopifyAddToCartOperation,
  ShopifyCartOperation,
  ShopifyCheckoutOperation,
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

/** Sentinel: cart service failed to load; UI should show "we'll display it later" instead of empty cart */
export const CART_UNAVAILABLE = Symbol('CART_UNAVAILABLE');
export type CartLoadResult = Cart | undefined | typeof CART_UNAVAILABLE;

/** Sentinel: goods/collections service failed to load; UI should show "we'll display it later" */
export const GOODS_UNAVAILABLE = Symbol('GOODS_UNAVAILABLE');
export type GoodsLoadResult<T> = T | typeof GOODS_UNAVAILABLE;

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
  const upstreamUnavailableMessage =
    'Service temporarily unavailable. Please try again in a moment.';

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

    const text = await result.text();
    let body: T & {
      error?: { message?: string };
      errors?: Array<{ message: string; cause?: unknown; status?: number }>;
    };

    try {
      body = JSON.parse(text) as typeof body;
    } catch {
      const rawMessage = text?.trim() || 'Invalid response from server';
      const message =
        rawMessage.toLowerCase().includes('no healthy upstream')
          ? upstreamUnavailableMessage
          : rawMessage;
      throw {
        cause: 'unknown',
        status: result.status || 500,
        message,
        query
      };
    }

    if (body.errors) {
      throw body.errors[0];
    }

    const bffMessage = body.error?.message ?? '';
    const isUpstreamUnavailable =
      bffMessage.toLowerCase().includes('no healthy upstream') ||
      bffMessage.toLowerCase().includes('failed to fetch from subgraph');
    if (body.error && isUpstreamUnavailable) {
      throw {
        cause: 'unknown',
        status: result.status || 500,
        message: upstreamUnavailableMessage,
        query
      };
    }
    if (body.error) {
      throw {
        cause: 'unknown',
        status: result.status || 500,
        message: bffMessage || 'Request failed',
        query
      };
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

    if (
      typeof e === 'object' &&
      e !== null &&
      'status' in e &&
      'message' in e &&
      'query' in e
    ) {
      throw e;
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
      const rawGood = Number.isFinite(numericGoodId) ? await getGood(numericGoodId) : undefined;
      const good = rawGood === GOODS_UNAVAILABLE ? undefined : rawGood;
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

export async function getCart(cartId: string | undefined): Promise<CartLoadResult> {
  if (!cartId) {
    return undefined;
  }

  try {
    const res = await shopifyFetch<ShopifyCartOperation>({
      query: getCartQuery,
      variables: {
        customerId: cartId
      },
      cache: 'no-store'
    });

    return buildCartFromState(res.body.data.getCart?.state, cartId);
  } catch {
    // Cart service unavailable — return sentinel so UI can show "we couldn't load your cart, we'll show it later"
    return CART_UNAVAILABLE;
  }
}

export async function checkout(input: CheckoutInput): Promise<CheckoutResult> {
  const res = await shopifyFetch<ShopifyCheckoutOperation>({
    query: checkoutMutation,
    variables: {
      input
    },
    cache: 'no-store'
  });

  return res.body.data.checkout;
}

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
  } catch {
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
  } catch {
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
      // Filter out the `hidden` collections.
      ...reshapeCollections(shopifyCollections).filter(
        (collection) => !collection.handle.startsWith('hidden')
      ),
    ];

    return collections;
  } catch {
    return GOODS_UNAVAILABLE;
  }
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
  } catch {
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
  } catch {
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
  } catch {
    return GOODS_UNAVAILABLE;
  }
}

