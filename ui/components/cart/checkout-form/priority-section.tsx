'use client';

import type { UseFormRegister } from 'react-hook-form';
import type { CheckoutFormInput } from '../checkout-form';

type PrioritySectionProps = {
  register: UseFormRegister<CheckoutFormInput>;
};

export function PrioritySection({ register }: PrioritySectionProps) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-[var(--color-foreground)]">Delivery Priority</h3>
      <div className="flex gap-4">
        <label className="flex cursor-pointer items-center">
          <input
            type="radio"
            value="NORMAL"
            {...register('priority')}
            className="h-4 w-4 text-[var(--color-primary)] focus:ring-[var(--color-primary)]"
          />
          <span className="ml-2 text-sm text-[var(--color-muted-foreground)]">Normal Delivery</span>
        </label>
        <label className="flex cursor-pointer items-center">
          <input
            type="radio"
            value="URGENT"
            {...register('priority')}
            className="h-4 w-4 text-[var(--color-primary)] focus:ring-[var(--color-primary)]"
          />
          <span className="ml-2 text-sm text-[var(--color-muted-foreground)]">
            Urgent Delivery (+$10)
          </span>
        </label>
      </div>
    </div>
  );
}
