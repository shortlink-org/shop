'use client';

import { DEFAULT_OPTION } from 'lib/constants';
import type { Cart, CartItem, Good, GoodVariant } from 'lib/shopify/types';
import { CART_UNAVAILABLE, type CartLoadResult } from 'lib/shopify';
import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';

type UpdateType = 'plus' | 'minus' | 'delete';

type CartAction =
  | { type: 'UPDATE_ITEM'; payload: { merchandiseId: string; updateType: UpdateType } }
  | { type: 'ADD_ITEM'; payload: { variant: GoodVariant; good: Good } }
  | { type: 'SET_CART_ID'; payload: { cartId: string } };

type CartContextType = {
  cart: Cart | undefined;
  /** True when cart service failed to load — show "we couldn't load your cart, we'll show it later" */
  cartUnavailable: boolean;
  updateCartItem: (merchandiseId: string, updateType: UpdateType) => void;
  addCartItem: (variant: GoodVariant, good: Good) => void;
  setCartId: (cartId: string) => void;
};

const CartContext = createContext<CartContextType | undefined>(undefined);

function calculateItemCost(quantity: number, price: number): number {
  return price * quantity;
}

function updateCartItem(item: CartItem, updateType: UpdateType): CartItem | null {
  if (updateType === 'delete') return null;

  const newQuantity = updateType === 'plus' ? item.quantity + 1 : item.quantity - 1;
  if (newQuantity === 0) return null;

  const singleItemAmount = Number(item.cost.totalAmount.amount) / item.quantity;
  const newTotalAmount = calculateItemCost(newQuantity, singleItemAmount);

  return {
    ...item,
    quantity: newQuantity,
    cost: {
      ...item.cost,
      totalAmount: {
        ...item.cost.totalAmount,
        amount: newTotalAmount
      }
    }
  };
}

function createOrUpdateCartItem(
  existingItem: CartItem | undefined,
  variant: GoodVariant,
  good: Good
): CartItem {
  const quantity = existingItem ? existingItem.quantity + 1 : 1;
  const totalAmount = calculateItemCost(quantity, variant.price.amount);

  const merchandiseTitle =
    variant.selectedOptions.length > 0 && variant.title ? variant.title : DEFAULT_OPTION;

  return {
    id: existingItem?.id,
    quantity,
    cost: {
      totalAmount: {
        amount: totalAmount,
        currencyCode: variant.price.currencyCode
      }
    },
    merchandise: {
      id: variant.id,
      title: merchandiseTitle,
      selectedOptions: variant.selectedOptions,
      product: {
        id: good.id,
        handle: good.name,
        title: good.name
      }
    }
  };
}

function updateCartTotals(lines: CartItem[]): Pick<Cart, 'totalQuantity' | 'cost'> {
  const totalQuantity = lines.reduce((sum, item) => sum + item.quantity, 0);
  const totalAmount = lines.reduce((sum, item) => sum + Number(item.cost.totalAmount.amount), 0);
  const currencyCode = lines[0]?.cost.totalAmount.currencyCode ?? 'USD';

  return {
    totalQuantity,
    cost: {
      subtotalAmount: { amount: totalAmount, currencyCode },
      totalAmount: { amount: totalAmount, currencyCode },
      totalTaxAmount: { amount: 0, currencyCode }
    }
  };
}

function createEmptyCart(): Cart {
  return {
    id: undefined,
    checkoutUrl: '',
    totalQuantity: 0,
    lines: [],
    cost: {
      subtotalAmount: { amount: 0, currencyCode: 'USD' },
      totalAmount: { amount: 0, currencyCode: 'USD' },
      totalTaxAmount: { amount: 0, currencyCode: 'USD' }
    }
  };
}

function cartReducer(state: Cart | undefined, action: CartAction): Cart {
  const currentCart = state || createEmptyCart();

  switch (action.type) {
    case 'UPDATE_ITEM': {
      const { merchandiseId, updateType } = action.payload;
      const updatedLines = currentCart.lines
        .map((item) =>
          item.merchandise.id === merchandiseId ? updateCartItem(item, updateType) : item
        )
        .filter(Boolean) as CartItem[];

      if (updatedLines.length === 0) {
        return {
          ...currentCart,
          lines: [],
          totalQuantity: 0,
          cost: {
            subtotalAmount: { amount: 0, currencyCode: 'USD' },
            totalAmount: { amount: 0, currencyCode: 'USD' },
            totalTaxAmount: { amount: 0, currencyCode: 'USD' }
          }
        };
      }

      return { ...currentCart, ...updateCartTotals(updatedLines), lines: updatedLines };
    }
    case 'ADD_ITEM': {
      const { variant, good } = action.payload;

      const existingItem = currentCart.lines.find((item) => item.merchandise.id === variant.id);
      const updatedItem = createOrUpdateCartItem(existingItem, variant, good);

      const updatedLines = existingItem
        ? currentCart.lines.map((item) => (item.merchandise.id === variant.id ? updatedItem : item))
        : [...currentCart.lines, updatedItem];

      return { ...currentCart, ...updateCartTotals(updatedLines), lines: updatedLines };
    }
    case 'SET_CART_ID': {
      const { cartId } = action.payload;
      return { ...currentCart, id: cartId };
    }
    default:
      return currentCart;
  }
}

export function CartProvider({
  children,
  initialCartResult
}: {
  children: React.ReactNode;
  initialCartResult: CartLoadResult;
}) {
  const initialCart = initialCartResult === CART_UNAVAILABLE ? undefined : initialCartResult;
  const [cart, setCart] = useState<Cart | undefined>(initialCart);
  const [cartUnavailable, setCartUnavailable] = useState(initialCartResult === CART_UNAVAILABLE);

  useEffect(() => {
    if (initialCartResult === CART_UNAVAILABLE) {
      setCartUnavailable(true);
      return;
    }

    setCartUnavailable(false);
    setCart(initialCartResult);
  }, [initialCartResult]);

  const updateCartItem = (merchandiseId: string, updateType: UpdateType) => {
    setCart((currentCart) =>
      cartReducer(currentCart, { type: 'UPDATE_ITEM', payload: { merchandiseId, updateType } })
    );
  };

  const addCartItem = (variant: GoodVariant, good: Good) => {
    setCart((currentCart) =>
      cartReducer(currentCart, { type: 'ADD_ITEM', payload: { variant, good } })
    );
  };

  const setCartId = (cartId: string) => {
    if (!cartId) return;
    setCart((currentCart) =>
      cartReducer(currentCart, { type: 'SET_CART_ID', payload: { cartId } })
    );
  };

  const value = useMemo(
    () => ({
      cart,
      cartUnavailable,
      updateCartItem,
      addCartItem,
      setCartId
    }),
    [cart, cartUnavailable]
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart() {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
}
