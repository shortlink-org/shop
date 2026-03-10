'use client';

import { useMemo, useState } from 'react';
import { toast } from 'sonner';
import { removeItem, updateItemQuantity } from './actions';
import { useCart } from './cart-context';
import { cartToBasketItems, formatCartMoney } from './ui-kit';

function getReverseUpdateType(updateType: 'plus' | 'minus') {
  return updateType === 'plus' ? 'minus' : 'plus';
}

export function useCartBasket() {
  const { cart, updateCartItem } = useCart();
  const [pendingItemIds, setPendingItemIds] = useState<Record<string, boolean>>({});

  const items = useMemo(() => cartToBasketItems(cart), [cart]);
  const subtotal = useMemo(
    () =>
      formatCartMoney(
        cart?.cost.subtotalAmount.amount ?? 0,
        cart?.cost.subtotalAmount.currencyCode ?? 'USD'
      ),
    [cart]
  );

  const setItemPending = (itemId: string, isPending: boolean) => {
    setPendingItemIds((prev) => {
      if (isPending) {
        return { ...prev, [itemId]: true };
      }

      const next = { ...prev };
      delete next[itemId];
      return next;
    });
  };

  const handleRemoveItem = async (itemId: number | string) => {
    const merchandiseId = String(itemId);

    if (pendingItemIds[merchandiseId]) {
      return;
    }

    setItemPending(merchandiseId, true);
    updateCartItem(merchandiseId, 'delete');

    try {
      const message = await removeItem(null, merchandiseId);
      if (message) {
        toast.error(message);
      }
    } catch {
      toast.error('Error removing item from cart');
    } finally {
      setItemPending(merchandiseId, false);
    }
  };

  const handleQuantityChange = async (itemId: number | string, quantity: number) => {
    const merchandiseId = String(itemId);
    const currentItem = cart?.lines.find((item) => item.merchandise.id === merchandiseId);

    if (
      !currentItem ||
      pendingItemIds[merchandiseId] ||
      quantity < 1 ||
      quantity === currentItem.quantity
    ) {
      return;
    }

    const updateType = quantity > currentItem.quantity ? 'plus' : 'minus';

    setItemPending(merchandiseId, true);
    updateCartItem(merchandiseId, updateType);

    try {
      const message = await updateItemQuantity(null, { merchandiseId, quantity });

      if (message) {
        updateCartItem(merchandiseId, getReverseUpdateType(updateType));
        toast.error(message);
      }
    } catch {
      updateCartItem(merchandiseId, getReverseUpdateType(updateType));
      toast.error('Error updating item quantity');
    } finally {
      setItemPending(merchandiseId, false);
    }
  };

  return {
    items,
    subtotal,
    handleRemoveItem,
    handleQuantityChange
  };
}
