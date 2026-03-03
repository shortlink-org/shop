'use client';

import { PlusIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';
import { addItem, type AddItemResult } from 'components/cart/actions';
import { Good, GoodVariant } from 'lib/shopify/types';
import { useFormState } from 'react-dom';
import { useCart } from './cart-context';

function SubmitButton({
  availableForSale,
  selectedVariantId
}: {
  availableForSale: boolean;
  selectedVariantId: string | undefined;
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
      aria-label="Add to cart"
      className={clsx(buttonClasses, {
        'hover:opacity-90': true
      })}
    >
      <div className="absolute left-0 ml-4">
        <PlusIcon className="h-5" />
      </div>
      Add To Cart
    </button>
  );
}

export function AddToCart({ good }: { good: Good }) {
  const { addCartItem } = useCart();
  const [result, formAction] = useFormState<AddItemResult | null, string | undefined>(
    addItem,
    null
  );
  const actionWithVariant = formAction.bind(null, good.id);
  const optimisticVariant: GoodVariant = {
    id: good.id,
    title: good.name,
    availableForSale: true,
    selectedOptions: [],
    price: { amount: good.price, currencyCode: 'USD' }
  };

  return (
    <form
      action={async () => {
        addCartItem(optimisticVariant, good);
        await actionWithVariant();
      }}
    >
      <SubmitButton availableForSale={true} selectedVariantId={good.id} />
      <p aria-live="polite" className="sr-only" role="status">
        {result?.ok === false ? result.message : ''}
      </p>
    </form>
  );
}
