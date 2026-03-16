'use client';

import { Button } from '@shortlink-org/ui-kit';
import type { UseFormRegister, FieldErrors } from 'react-hook-form';
import type { CheckoutFormInput } from '../checkout-form';

type DeliveryAddressSectionProps = {
  register: UseFormRegister<CheckoutFormInput>;
  errors: FieldErrors<CheckoutFormInput>;
  onRandomAddress: () => void;
  randomAddressLoading: boolean;
};

export function DeliveryAddressSection({
  register,
  errors,
  onRandomAddress,
  randomAddressLoading
}: DeliveryAddressSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-[var(--color-foreground)]">Delivery Address</h3>
        <Button type="button" variant="secondary" onClick={onRandomAddress} loading={randomAddressLoading}>
          Random address
        </Button>
      </div>

      <div>
        <label htmlFor="street" className="block text-sm font-medium text-[var(--color-muted-foreground)]">
          Street Address *
        </label>
        <input
          type="text"
          id="street"
          {...register('deliveryAddress.street')}
          className={`mt-1 block w-full rounded-md border bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none ${
            errors.deliveryAddress?.street ? 'border-[var(--color-destructive)]' : 'border-[var(--color-border)]'
          }`}
          placeholder="123 Main Street, Apt 4"
        />
        {errors.deliveryAddress?.street && (
          <p className="mt-1 text-sm text-[var(--color-destructive)]">
            {errors.deliveryAddress.street.message}
          </p>
        )}
      </div>

      <div>
        <label htmlFor="city" className="block text-sm font-medium text-[var(--color-muted-foreground)]">
          City *
        </label>
        <input
          type="text"
          id="city"
          {...register('deliveryAddress.city')}
          className={`mt-1 block w-full rounded-md border bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none ${
            errors.deliveryAddress?.city ? 'border-[var(--color-destructive)]' : 'border-[var(--color-border)]'
          }`}
          placeholder="Berlin"
        />
        {errors.deliveryAddress?.city && (
          <p className="mt-1 text-sm text-[var(--color-destructive)]">
            {errors.deliveryAddress.city.message}
          </p>
        )}
      </div>

      <div>
        <label htmlFor="country" className="block text-sm font-medium text-[var(--color-muted-foreground)]">
          Country *
        </label>
        <select
          id="country"
          {...register('deliveryAddress.country')}
          className={`mt-1 block w-full rounded-md border bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none ${
            errors.deliveryAddress?.country ? 'border-[var(--color-destructive)]' : 'border-[var(--color-border)]'
          }`}
        >
          <option value="Germany">Germany</option>
          <option value="Austria">Austria</option>
          <option value="Switzerland">Switzerland</option>
          <option value="Netherlands">Netherlands</option>
          <option value="Belgium">Belgium</option>
          <option value="France">France</option>
        </select>
        {errors.deliveryAddress?.country && (
          <p className="mt-1 text-sm text-[var(--color-destructive)]">
            {errors.deliveryAddress.country.message}
          </p>
        )}
      </div>
    </div>
  );
}
