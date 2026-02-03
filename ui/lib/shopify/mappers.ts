import type {
  Cart,
  Collection,
  Good,
  ShopifyCartOperation,
  ShopifyCollection,
  ShopifyProduct
} from './types';

const DEFAULT_CURRENCY = 'USD';

export function normalizeGood(good: ShopifyProduct): Good {
  const createdAt = good.createdAt ?? good.created_at ?? new Date().toISOString();
  const updatedAt = good.updatedAt ?? good.updated_at ?? new Date().toISOString();

  return {
    ...good,
    createdAt,
    updatedAt
  };
}

export function createEmptyCart(cartId?: string): Cart {
  return {
    id: cartId,
    checkoutUrl: '',
    totalQuantity: 0,
    lines: [],
    cost: {
      subtotalAmount: { amount: 0, currencyCode: DEFAULT_CURRENCY },
      totalAmount: { amount: 0, currencyCode: DEFAULT_CURRENCY },
      totalTaxAmount: { amount: 0, currencyCode: DEFAULT_CURRENCY }
    }
  };
}

export function reshapeCollection(collection: ShopifyCollection): Collection | undefined {
  if (!collection) {
    return undefined;
  }

  return {
    ...collection,
    path: `/search/${collection.handle}`
  };
}

export function reshapeCollections(collections: ShopifyCollection[]): Collection[] {
  const reshapedCollections: Collection[] = [];

  for (const collection of collections) {
    if (collection) {
      const reshapedCollection = reshapeCollection(collection);

      if (reshapedCollection) {
        reshapedCollections.push(reshapedCollection);
      }
    }
  }

  return reshapedCollections;
}

export { DEFAULT_CURRENCY };

export type CartState = ShopifyCartOperation['data']['getCart']['state'];
