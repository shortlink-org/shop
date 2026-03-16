'use client';

import type { UseFormRegister, UseFormSetValue, FieldErrors } from 'react-hook-form';
import type { CheckoutFormInput } from '../checkout-form';

const TIME_SLOTS = [
  { label: '09:00 - 12:00', start: '09:00', end: '12:00' },
  { label: '12:00 - 15:00', start: '12:00', end: '15:00' },
  { label: '15:00 - 18:00', start: '15:00', end: '18:00' },
  { label: '18:00 - 21:00', start: '18:00', end: '21:00' }
] as const;

type DeliveryPeriodSectionProps = {
  register: UseFormRegister<CheckoutFormInput>;
  setValue: UseFormSetValue<CheckoutFormInput>;
  selectedTimeSlot: string;
  selectedDeliveryDate: string;
  errors: FieldErrors<CheckoutFormInput>;
  tomorrowDate: string;
  maxDeliveryDate: string;
};

export function DeliveryPeriodSection({
  register,
  setValue,
  selectedTimeSlot,
  selectedDeliveryDate,
  errors,
  tomorrowDate,
  maxDeliveryDate
}: DeliveryPeriodSectionProps) {

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-[var(--color-foreground)]">Delivery Time</h3>

      <div>
        <div className="flex items-center justify-between">
          <label
            htmlFor="deliveryDate"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Delivery Date *
          </label>
          <button
            type="button"
            onClick={() =>
              setValue('deliveryDate', tomorrowDate, {
                shouldDirty: true,
                shouldTouch: true,
                shouldValidate: true
              })
            }
            className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
              selectedDeliveryDate === tomorrowDate
                ? 'border-[var(--color-primary)] bg-[var(--color-primary)] text-white'
                : 'border-[var(--color-border)] text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)]'
            }`}
          >
            Tomorrow
          </button>
        </div>
        <input
          type="date"
          id="deliveryDate"
          {...register('deliveryDate')}
          min={tomorrowDate}
          max={maxDeliveryDate}
          className={`mt-1 block w-full rounded-md border bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none ${
            errors.deliveryDate ? 'border-[var(--color-destructive)]' : 'border-[var(--color-border)]'
          }`}
        />
        {errors.deliveryDate && (
          <p className="mt-1 text-sm text-[var(--color-destructive)]">{errors.deliveryDate.message}</p>
        )}
      </div>

      <div>
        <fieldset>
          <legend className="block text-sm font-medium text-[var(--color-muted-foreground)]">
            Time Slot *
          </legend>
          <div className="mt-2 grid grid-cols-2 gap-2">
            {TIME_SLOTS.map((slot) => (
              <button
                key={slot.label}
                type="button"
                onClick={() => setValue('selectedTimeSlot', slot.label, { shouldValidate: true })}
                className={`rounded-md border px-4 py-2 text-sm font-medium transition-colors ${
                  selectedTimeSlot === slot.label
                    ? 'border-sky-600 bg-sky-600 text-white shadow-sm hover:bg-sky-700'
                    : 'border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)]'
                }`}
              >
                {slot.label}
              </button>
            ))}
          </div>
        </fieldset>
        {errors.selectedTimeSlot && (
          <p className="mt-1 text-sm text-[var(--color-destructive)]">
            {errors.selectedTimeSlot.message}
          </p>
        )}
      </div>
    </div>
  );
}
