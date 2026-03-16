'use client';

import clsx from 'clsx';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { removeItem } from 'components/cart/actions';
import type { CartItem } from 'lib/shopify/types';
import { useActionState } from 'react';

type UpdateType = 'plus' | 'minus' | 'delete';

export function DeleteItemButton({
  item,
  optimisticUpdate
}: {
  item: CartItem;
  optimisticUpdate: (merchandiseId: string, updateType: UpdateType) => void;
}) {
  const [message, formAction, isPending] = useActionState(removeItem, null);
  const merchandiseId = item.merchandise.id;
  const actionWithVariant = formAction.bind(null, merchandiseId);

  return (
    <form
      action={async () => {
        optimisticUpdate(merchandiseId, 'delete');
        await actionWithVariant();
      }}
    >
      <button
        type="submit"
        aria-label="Remove cart item"
        disabled={isPending}
        className="flex h-[24px] w-[24px] items-center justify-center rounded-full bg-neutral-500 disabled:opacity-60 disabled:cursor-not-allowed"
      >
        <XMarkIcon
          className={clsx(
            'mx-[1px] h-4 w-4 text-white dark:text-black',
            isPending && 'animate-pulse'
          )}
        />
      </button>
      <p aria-live="polite" className="sr-only" role="status">
        {message}
      </p>
    </form>
  );
}
