'use server';

import { TAGS } from 'lib/constants';
import { addToCart, checkout, shopifyFetch, updateCart } from 'lib/shopify';
import type { CheckoutInput } from 'lib/shopify/types';
import { revalidateTag } from 'next/cache';
import { cookies, headers } from 'next/headers';
import { redirect } from 'next/navigation';

export type AddItemResult = { ok: true; cartId: string } | { ok: false; message: string };
export type CheckoutActionResult =
  | { ok: true; orderId: string }
  | { ok: false; message: string };
export type RandomAddressResult =
  | {
      ok: true;
      address: { street?: string; city?: string; postalCode?: string; country?: string };
    }
  | { ok: false; message: string };

type RandomAddressOperation = {
  data?: {
    randomAddress?: {
      address?: { street?: string; city?: string; postalCode?: string; country?: string };
    };
  };
};

async function getHeaderCustomerId() {
  const requestHeaders = await headers();
  return requestHeaders.get('x-user-id') ?? requestHeaders.get('x-customer-id') ?? undefined;
}

async function getExistingCustomerId() {
  const cookieStore = await cookies();
  const headerCustomerId = await getHeaderCustomerId();
  if (headerCustomerId) {
    cookieStore.set('cartId', headerCustomerId);
    return headerCustomerId;
  }

  return '';
}

async function getOrCreateCartId() {
  const existingCustomerId = await getExistingCustomerId();
  return existingCustomerId;
}

export async function addItem(
  prevState: unknown,
  selectedVariantId: string | undefined
): Promise<AddItemResult> {
  if (!selectedVariantId) {
    console.error('[addItem] missing selectedVariantId');
    return { ok: false, message: 'Unable to add this product right now.' };
  }

  const cartId = await getOrCreateCartId();
  if (!cartId) {
    console.error('[addItem] getOrCreateCartId returned empty');
    return { ok: false, message: 'Unable to create a cart. Please try again.' };
  }

  const authHeader = (await headers()).get('authorization') ?? undefined;
  const userId = await getHeaderCustomerId();

  try {
    await addToCart([{ goodId: selectedVariantId, quantity: 1 }], {
      authorization: authHeader,
      userId
    });
    revalidateTag(TAGS.cart, 'max');
    return { ok: true, cartId };
  } catch (e) {
    const err = e as Record<string, unknown>;
    const msgRaw = err?.message ?? (e instanceof Error ? e.message : null);
    const message =
      typeof msgRaw === 'string'
        ? msgRaw
        : msgRaw != null
          ? JSON.stringify(msgRaw)
          : err?.error != null &&
              typeof (err.error as Record<string, unknown>)?.message === 'string'
            ? (err.error as { message: string }).message
            : null;
    const serialized =
      typeof e === 'object' && e !== null
        ? JSON.stringify(
            e,
            (_, v) => (v instanceof Error ? { name: v.name, message: v.message } : v),
            2
          )
        : String(e);
    console.error('[addItem] addToCart failed', {
      cartId,
      selectedVariantId,
      message: message ?? '(no message)',
      status: err?.status,
      serialized
    });
    return {
      ok: false,
      message: typeof message === 'string' && message.trim() ? message : 'Error adding item to cart'
    };
  }
}

export async function removeItem(prevState: any, merchandiseId: string) {
  const cartId = await getExistingCustomerId();

  if (!cartId || !merchandiseId) {
    return 'Missing cart ID';
  }

  const authHeader = (await headers()).get('authorization') ?? undefined;
  const userId = await getHeaderCustomerId();

  try {
    await updateCart([{ id: merchandiseId, merchandiseId, quantity: 0 }], {
      authorization: authHeader,
      userId
    });
    revalidateTag(TAGS.cart, 'max');
  } catch (e) {
    return 'Error removing item from cart';
  }
}

export async function updateItemQuantity(
  prevState: any,
  payload: {
    merchandiseId: string;
    quantity: number;
  }
) {
  const cartId = await getExistingCustomerId();

  if (!cartId) {
    return 'Missing cart ID';
  }

  const { merchandiseId, quantity } = payload;
  const authHeader = (await headers()).get('authorization') ?? undefined;
  const userId = await getHeaderCustomerId();

  try {
    await updateCart([{ id: merchandiseId, merchandiseId, quantity }], {
      authorization: authHeader,
      userId
    });
    revalidateTag(TAGS.cart, 'max');
  } catch (e) {
    console.error(e);
    return 'Error updating item quantity';
  }
}

export async function redirectToCheckout(): Promise<void> {
  const cartId = await getExistingCustomerId();

  if (!cartId) {
    console.error('Missing cart ID');
    return;
  }

  redirect('/checkout');
}

export async function submitCheckout(input: CheckoutInput): Promise<CheckoutActionResult> {
  const authHeader = (await headers()).get('authorization') ?? undefined;
  const userId = await getHeaderCustomerId();

  try {
    const result = await checkout(input, {
      authorization: authHeader,
      userId
    });

    if (!result.orderId) {
      return { ok: false, message: 'Failed to create order. Please try again.' };
    }

    revalidateTag(TAGS.cart, 'max');
    return { ok: true, orderId: result.orderId };
  } catch (e) {
    const err = e as Record<string, unknown>;
    const msgRaw = err?.message ?? (e instanceof Error ? e.message : null);
    const message =
      typeof msgRaw === 'string'
        ? msgRaw.trim()
        : msgRaw != null
          ? JSON.stringify(msgRaw)
          : err?.error != null && typeof (err.error as Record<string, unknown>)?.message === 'string'
            ? String((err.error as { message: string }).message).trim()
            : null;
    const serialized =
      typeof e === 'object' && e !== null
        ? JSON.stringify(
            e,
            (_, v) => (v instanceof Error ? { name: v.name, message: v.message } : v),
            2
          )
        : String(e);

    console.error('[submitCheckout] checkout failed', {
      hasAuthorization: Boolean(authHeader),
      userId,
      message: message ?? '(no message)',
      status: err?.status,
      serialized
    });

    return {
      ok: false,
      message:
        typeof message === 'string' && message.length > 0
          ? message
          : 'An error occurred during checkout. Please try again.'
    };
  }
}

export async function fetchRandomAddress(): Promise<RandomAddressResult> {
  const authHeader = (await headers()).get('authorization') ?? undefined;

  try {
    const res = await shopifyFetch<RandomAddressOperation>({
      query: `
        query GetRandomAddress {
          randomAddress {
            address {
              street
              city
              postalCode
              country
            }
          }
        }
      `,
      cache: 'no-store',
      headers: authHeader ? { Authorization: authHeader } : undefined
    });

    const address = res.body.data?.randomAddress?.address;
    if (!address) {
      return { ok: false, message: 'Unable to load a random address right now.' };
    }

    return { ok: true, address };
  } catch (e) {
    const err = e as Record<string, unknown>;
    const message =
      typeof err?.message === 'string' && err.message.trim()
        ? err.message.trim()
        : 'Unable to load a random address right now.';

    console.error('[fetchRandomAddress] failed', {
      hasAuthorization: Boolean(authHeader),
      message,
      status: err?.status
    });

    return { ok: false, message };
  }
}

export async function getCheckoutCartId(): Promise<string> {
  return (await getExistingCustomerId()) || '';
}

export async function createCartAndSetCookie() {
  const customerId = await getHeaderCustomerId();
  if (!customerId) {
    return;
  }

  const cookieStore = await cookies();
  cookieStore.set('cartId', customerId);
}
