import {
  addToCartMutation,
  removeFromCartMutation,
  checkoutMutation
} from '../mutations/cart';
import { getCartQuery } from '../queries/cart';
import { shopifyFetch } from '../fetch';
import { createEmptyCart, DEFAULT_CURRENCY } from '../mappers';
import type { CartState } from '../mappers';
import { CART_UNAVAILABLE } from '../sentinels';
import type { CartLoadResult } from '../sentinels';
import type {
  Cart,
  CheckoutInput,
  CheckoutResult,
  ShopifyAddToCartOperation,
  ShopifyCartOperation,
  ShopifyCheckoutOperation,
  ShopifyRemoveFromCartOperation
} from '../types';
import { getGood } from './good';
import { GOODS_UNAVAILABLE as GOODS_UNAVAILABLE_SENTINEL } from '../sentinels';

function buildCartFromState(
  state: CartState | undefined | null,
  fallbackId?: string,
  options?: RequestOptions
): Promise<Cart> {
  const items = state?.items ?? [];

  if (!items.length) {
    return Promise.resolve(createEmptyCart(state?.cartId ?? fallbackId));
  }

  return Promise.all(
    items.map(async (item) => {
      const goodId = item?.goodId ?? '';
      const quantity = item?.quantity ?? 0;

      if (!goodId || quantity <= 0) {
        return null;
      }

      const numericGoodId = Number(goodId);
      const rawGood = Number.isFinite(numericGoodId)
        ? await getGood(numericGoodId, options)
        : undefined;
      const good = rawGood === GOODS_UNAVAILABLE_SENTINEL ? undefined : rawGood;
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
  ).then((lineItems) => {
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
  });
}

export async function createCart(): Promise<Cart> {
  const cartId = crypto.randomUUID();
  return createEmptyCart(cartId);
}

export type RequestOptions = { authorization?: string };

export async function addToCart(
  customerId: string,
  items: { goodId: string; quantity: number }[],
  options?: RequestOptions
): Promise<void> {
  await shopifyFetch<ShopifyAddToCartOperation>({
    query: addToCartMutation,
    variables: {
      addRequest: {
        customerId,
        items
      }
    },
    cache: 'no-store',
    headers: options?.authorization ? { Authorization: options.authorization } : {}
  });
}

export async function removeFromCart(
  customerId: string,
  items: { goodId: string; quantity: number }[],
  options?: RequestOptions
): Promise<void> {
  await shopifyFetch<ShopifyRemoveFromCartOperation>({
    query: removeFromCartMutation,
    variables: {
      removeRequest: {
        customerId,
        items
      }
    },
    cache: 'no-store',
    headers: options?.authorization ? { Authorization: options.authorization } : {}
  });
}

export async function updateCart(
  customerId: string,
  lines: { id: string; merchandiseId: string; quantity: number }[],
  options?: RequestOptions
): Promise<void> {
  const res = await shopifyFetch<ShopifyCartOperation>({
    query: getCartQuery,
    variables: {
      customerId
    },
    cache: 'no-store',
    headers: options?.authorization ? { Authorization: options.authorization } : {}
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
    await addToCart(customerId, itemsToAdd, options);
  }

  if (itemsToRemove.length > 0) {
    await removeFromCart(customerId, itemsToRemove, options);
  }
}

export async function getCart(
  cartId: string | undefined,
  options?: RequestOptions
): Promise<CartLoadResult> {
  if (!cartId) {
    return undefined;
  }

  try {
    const res = await shopifyFetch<ShopifyCartOperation>({
      query: getCartQuery,
      variables: {
        customerId: cartId
      },
      cache: 'no-store',
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    return buildCartFromState(res.body.data.getCart?.state, cartId, options);
  } catch {
    return CART_UNAVAILABLE;
  }
}

export async function checkout(
  input: CheckoutInput,
  options?: RequestOptions
): Promise<CheckoutResult> {
  const res = await shopifyFetch<ShopifyCheckoutOperation>({
    query: checkoutMutation,
    variables: {
      input
    },
    cache: 'no-store',
    headers: options?.authorization ? { Authorization: options.authorization } : {}
  });

  return res.body.data.checkout;
}
