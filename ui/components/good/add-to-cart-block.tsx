'use client';

import { AddToCartButton } from '@shortlink-org/ui-kit';
import { addItem } from 'components/cart/actions';
import { useCart } from 'components/cart/cart-context';
import { Good, GoodVariant } from 'lib/shopify/types';

export function AddToCartBlock({ good }: { good: Good }) {
  const { addCartItem } = useCart();
  const optimisticVariant: GoodVariant = {
    id: String(good.id),
    title: good.name,
    availableForSale: true,
    selectedOptions: [],
    price: { amount: good.price, currencyCode: 'USD' },
  };

  const handleAddToCart = async () => {
    addCartItem(optimisticVariant, good);
    await addItem(null, String(good.id));
  };

  return (
    <AddToCartButton
      text="Add to cart"
      ariaLabel="Add to cart"
      onAddToCart={handleAddToCart}
    />
  );
}
