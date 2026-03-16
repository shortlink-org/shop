'use client';

import { PlusIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';
import { addItem, type AddItemResult } from 'components/cart/actions';
import { DEFAULT_OPTION } from 'lib/constants';
import { Good, GoodVariant } from 'lib/shopify/types';
import { useEffect, useActionState } from 'react';
import { useCart } from './cart-context';

function SubmitButton({
  availableForSale,
  selectedVariantId,
  isPending
}: {
  availableForSale: boolean;
  selectedVariantId: string | undefined;
  isPending: boolean;
}) {
  const buttonClasses =
    'relative flex w-full items-center justify-center rounded-full bg-blue-600 p-4 tracking-wide text-white';
  const disabledClasses = 'cursor-not-allowed opacity-60 hover:opacity-60';

  if (!availableForSale) {
    return (
      <button disabled className={clsx(buttonClasses, disabledClasses)}>
        Out Of Stock
      </button>
    );
  }

  if (!selectedVariantId) {
    return (
      <button
        aria-label="Please select an option"
        disabled
        className={clsx(buttonClasses, disabledClasses)}
      >
        <div className="absolute left-0 ml-4">
          <PlusIcon className="h-5" />
        </div>
        Add To Cart
      </button>
    );
  }

  return (
    <button
      type="submit"
      aria-label="Add to cart"
      disabled={isPending}
      className={clsx(buttonClasses, {
        'hover:opacity-90': !isPending,
        [disabledClasses]: isPending
      })}
    >
      <div className="absolute left-0 ml-4">
        <PlusIcon className={clsx('h-5', isPending && 'animate-pulse')} />
      </div>
      {isPending ? 'Adding…' : 'Add To Cart'}
    </button>
  );
}

export function AddToCart({ good }: { good: Good }) {
  const { addCartItem, setCartId } = useCart();
  const [result, formAction, isPending] = useActionState<AddItemResult | null, string | undefined>(
    addItem,
    null
  );
  const actionWithVariant = formAction.bind(null, good.id);
  const optimisticVariant: GoodVariant = {
    id: good.id,
    title: DEFAULT_OPTION,
    availableForSale: true,
    selectedOptions: [],
    price: { amount: good.price, currencyCode: 'USD' }
  };

  useEffect(() => {
    if (result?.ok) {
      queueMicrotask(() => setCartId(result.cartId));
    }
  }, [result, setCartId]);

  return (
    <form
      action={async () => {
        addCartItem(optimisticVariant, good);
        await actionWithVariant();
      }}
    >
      <SubmitButton
        availableForSale={true}
        selectedVariantId={good.id}
        isPending={isPending}
      />
      <p aria-live="polite" className="sr-only" role="status">
        {result?.ok === false ? result.message : ''}
      </p>
    </form>
  );
}
