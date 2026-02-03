import type { Cart } from './types';

/** Sentinel: cart service failed to load; UI should show "we'll display it later" instead of empty cart. Uses Symbol.for so it can be passed to Client Components. */
export const CART_UNAVAILABLE = Symbol.for('CART_UNAVAILABLE');
export type CartLoadResult = Cart | undefined | typeof CART_UNAVAILABLE;

/** Sentinel: goods/collections service failed to load; UI should show "we'll display it later". Uses Symbol.for so it can be passed to Client Components. */
export const GOODS_UNAVAILABLE = Symbol.for('GOODS_UNAVAILABLE');
export type GoodsLoadResult<T> = T | typeof GOODS_UNAVAILABLE;
