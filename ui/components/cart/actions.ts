'use server';

import { TAGS } from 'lib/constants';
import { addToCart, createCart, updateCart } from 'lib/shopify';
import { revalidateTag } from 'next/cache';
import { cookies, headers } from 'next/headers';
import { redirect } from 'next/navigation';

export type AddItemResult = { ok: true } | { ok: false; message: string };

async function getHeaderCustomerId() {
  const requestHeaders = await headers();
  return requestHeaders.get('x-user-id') ?? requestHeaders.get('x-customer-id') ?? undefined;
}

async function getExistingCustomerId() {
  const cookieStore = await cookies();
  const existingCartId = cookieStore.get('cartId')?.value;

  if (existingCartId) {
    return existingCartId;
  }

  const headerCustomerId = await getHeaderCustomerId();
  if (headerCustomerId) {
    cookieStore.set('cartId', headerCustomerId);
    return headerCustomerId;
  }

  return '';
}

async function getOrCreateCartId() {
  const cookieStore = await cookies();
  const existingCustomerId = await getExistingCustomerId();
  if (existingCustomerId) return existingCustomerId;

  const newCart = await createCart();
  if (newCart.id) {
    cookieStore.set('cartId', newCart.id);
  }

  return newCart.id ?? '';
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

  try {
    await addToCart(cartId, [{ goodId: selectedVariantId, quantity: 1 }], {
      authorization: authHeader
    });
    revalidateTag(TAGS.cart, 'max');
    return { ok: true };
  } catch (e) {
    const err = e as Record<string, unknown>;
    const msgRaw = err?.message ?? (e instanceof Error ? e.message : null);
    const message =
      typeof msgRaw === 'string'
        ? msgRaw
        : msgRaw != null
          ? JSON.stringify(msgRaw)
          : (err?.error != null && typeof (err.error as Record<string, unknown>)?.message === 'string')
            ? (err.error as { message: string }).message
            : null;
    const serialized =
      typeof e === 'object' && e !== null
        ? JSON.stringify(e, (_, v) => (v instanceof Error ? { name: v.name, message: v.message } : v), 2)
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
      message: typeof message === 'string' && message.trim()
        ? message
        : 'Error adding item to cart'
    };
  }
}

export async function removeItem(prevState: any, merchandiseId: string) {
  const cartId = await getExistingCustomerId();

  if (!cartId || !merchandiseId) {
    return 'Missing cart ID';
  }

  const authHeader = (await headers()).get('authorization') ?? undefined;

  try {
    await updateCart(
      cartId,
      [{ id: merchandiseId, merchandiseId, quantity: 0 }],
      { authorization: authHeader }
    );
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

  try {
    await updateCart(cartId, [{ id: merchandiseId, merchandiseId, quantity }], {
      authorization: authHeader
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

export async function getCheckoutCartId(): Promise<string> {
  return (await getExistingCustomerId()) || '';
}

export async function createCartAndSetCookie() {
  const existingCartId = await getExistingCustomerId();

  if (existingCartId) {
    return;
  }

  const cookieStore = await cookies();
  const cart = await createCart();
  if (cart.id) {
    cookieStore.set('cartId', cart.id);
  }
}
