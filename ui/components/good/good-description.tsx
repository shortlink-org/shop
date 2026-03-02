'use client';

import { AddToCartButton, ProductDescription } from '@shortlink-org/ui-kit';
import { addItem } from 'components/cart/actions';
import { useCart } from 'components/cart/cart-context';
import Price from 'components/price';
import { Good, GoodVariant } from 'lib/shopify/types';

export function GoodDescription({ good }: { good: Good }) {
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
    <>
      <div className="mb-6 flex flex-col border-b pb-6 dark:border-neutral-700">
        <h1 className="mb-2 text-5xl font-medium">{good.name}</h1>
        <div className="mr-auto w-auto rounded-full bg-blue-600 p-2 text-sm text-white">
          <Price amount={good.price} />
        </div>
      </div>
      <ProductDescription
        description={good.description ?? undefined}
        highlights={[]}
        details={undefined}
      />
      <AddToCartButton
        text="Add to cart"
        ariaLabel="Add to cart"
        onAddToCart={handleAddToCart}
      />
    </>
  );
}
