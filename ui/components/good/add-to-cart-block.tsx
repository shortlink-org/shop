'use client';

import { AddToCartButton } from '@shortlink-org/ui-kit';
import { addItem } from 'components/cart/actions';
import { useCart } from 'components/cart/cart-context';
import { DEFAULT_OPTION } from 'lib/constants';
import { Good, GoodVariant } from 'lib/shopify/types';
import { useState } from 'react';
import { toast } from 'sonner';

export function AddToCartBlock({ good }: { good: Good }) {
  const { addCartItem, updateCartItem, setCartId } = useCart();
  const [isAdding, setIsAdding] = useState(false);

  const optimisticVariant: GoodVariant = {
    id: good.id,
    title: DEFAULT_OPTION,
    availableForSale: true,
    selectedOptions: [],
    price: { amount: good.price, currencyCode: 'USD' }
  };

  const handleAddToCart = async () => {
    if (isAdding) return;

    setIsAdding(true);
    addCartItem(optimisticVariant, good);
    try {
      const result = await addItem(null, good.id);
      if (!result.ok) {
        updateCartItem(good.id, 'minus');
        toast.error(result.message);
        return;
      }
      setCartId(result.cartId);
      toast.success(`${good.name} added to cart`);
    } catch {
      updateCartItem(good.id, 'minus');
      toast.error('Error adding item to cart');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <AddToCartButton
      text={isAdding ? 'Adding...' : 'Add to cart'}
      ariaLabel="Add to cart"
      onAddToCart={handleAddToCart}
      className={isAdding ? 'pointer-events-none opacity-70' : undefined}
    />
  );
}
