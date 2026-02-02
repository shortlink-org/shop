'use server';

import { TAGS } from 'lib/constants';
import { addToCart, createCart, updateCart } from 'lib/shopify';
import { revalidateTag } from 'next/cache';
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

async function getOrCreateCartId() {
  const cookieStore = await cookies();
  const existingCartId = cookieStore.get('cartId')?.value;

  if (existingCartId) {
    return existingCartId;
  }

  const newCart = await createCart();
  if (newCart.id) {
    cookieStore.set('cartId', newCart.id);
  }

  return newCart.id ?? '';
}

export async function addItem(prevState: any, selectedVariantId: string | undefined) {
  if (!selectedVariantId) {
    return 'Error adding item to cart';
  }

  const cartId = await getOrCreateCartId();
  if (!cartId) {
    return 'Missing cart ID';
  }

  try {
    await addToCart(cartId, [{ goodId: selectedVariantId, quantity: 1 }]);
    revalidateTag(TAGS.cart, 'max');
  } catch (e) {
    return 'Error adding item to cart';
  }
}

export async function removeItem(prevState: any, merchandiseId: string) {
  const cartId = (await cookies()).get('cartId')?.value;

  if (!cartId || !merchandiseId) {
    return 'Missing cart ID';
  }

  try {
    await updateCart(cartId, [{ id: merchandiseId, merchandiseId, quantity: 0 }]);
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
  const cartId = (await cookies()).get('cartId')?.value;

  if (!cartId) {
    return 'Missing cart ID';
  }

  const { merchandiseId, quantity } = payload;

  try {
    await updateCart(cartId, [{ id: merchandiseId, merchandiseId, quantity }]);
    revalidateTag(TAGS.cart, 'max');
  } catch (e) {
    console.error(e);
    return 'Error updating item quantity';
  }
}

export async function redirectToCheckout(): Promise<void> {
  const cartId = (await cookies()).get('cartId')?.value;

  if (!cartId) {
    console.error('Missing cart ID');
    return;
  }

  redirect('/checkout');
}

export async function createCartAndSetCookie() {
  const cookieStore = await cookies();
  const existingCartId = cookieStore.get('cartId')?.value;

  if (existingCartId) {
    return;
  }

  const cart = await createCart();
  if (cart.id) {
    cookieStore.set('cartId', cart.id);
  }
}
