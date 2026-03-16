'use client';

import { MinusIcon, PlusIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';
import { updateItemQuantity } from 'components/cart/actions';
import type { CartItem } from 'lib/shopify/types';
import { useActionState } from 'react';

type UpdateType = 'plus' | 'minus' | 'delete';

function SubmitButton({
  type,
  isPending
}: {
  type: 'plus' | 'minus';
  isPending: boolean;
}) {
  return (
    <button
      type="submit"
      aria-label={type === 'plus' ? 'Increase item quantity' : 'Reduce item quantity'}
      disabled={isPending}
      className={clsx(
        'ease flex h-full max-w-[36px] min-w-[36px] flex-none items-center justify-center rounded-full p-2 transition-all duration-200 hover:border-neutral-800 hover:opacity-80',
        {
          'ml-auto': type === 'minus',
          'opacity-60 cursor-not-allowed': isPending
        }
      )}
    >
      {type === 'plus' ? (
        <PlusIcon className={clsx('h-4 w-4 dark:text-neutral-500', isPending && 'animate-pulse')} />
      ) : (
        <MinusIcon className={clsx('h-4 w-4 dark:text-neutral-500', isPending && 'animate-pulse')} />
      )}
    </button>
  );
}

export function EditItemQuantityButton({
  item,
  type,
  optimisticUpdate
}: {
  item: CartItem;
  type: 'plus' | 'minus';
  optimisticUpdate: (merchandiseId: string, updateType: UpdateType) => void;
}) {
  const [message, formAction, isPending] = useActionState(updateItemQuantity, null);
  const payload = {
    merchandiseId: item.merchandise.id,
    quantity: type === 'plus' ? item.quantity + 1 : item.quantity - 1
  };
  const actionWithVariant = formAction.bind(null, payload);

  return (
    <form
      action={async () => {
        optimisticUpdate(payload.merchandiseId, type);
        await actionWithVariant();
      }}
    >
      <SubmitButton type={type} isPending={isPending} />
      <p aria-live="polite" className="sr-only" role="status">
        {message}
      </p>
    </form>
  );
}
