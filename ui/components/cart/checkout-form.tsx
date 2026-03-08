'use client';

import { Button } from '@shortlink-org/ui-kit';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useSession } from '@/contexts/SessionContext';
import { useEffect } from 'react';

const TIME_SLOTS = [
  { label: '09:00 - 12:00', start: '09:00', end: '12:00' },
  { label: '12:00 - 15:00', start: '12:00', end: '15:00' },
  { label: '15:00 - 18:00', start: '15:00', end: '18:00' },
  { label: '18:00 - 21:00', start: '18:00', end: '21:00' }
] as const;

const TIME_SLOT_LABELS = TIME_SLOTS.map((s) => s.label);
const COUNTRY_VALUES = [
  'Germany',
  'Austria',
  'Switzerland',
  'Netherlands',
  'Belgium',
  'France'
] as const;

function getMinDate(): string {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  return tomorrow.toISOString().split('T')[0] ?? '';
}

function getMaxDate(): string {
  const maxDate = new Date();
  maxDate.setDate(maxDate.getDate() + 14);
  return maxDate.toISOString().split('T')[0] ?? '';
}

const countrySchema = z.enum(COUNTRY_VALUES);
const prioritySchema = z.enum(['NORMAL', 'URGENT']);
const timeSlotLabelSchema = z.enum(TIME_SLOT_LABELS as unknown as [string, ...string[]]);

const checkoutFormInputSchema = z
  .object({
    deliveryAddress: z.object({
      street: z.string().min(1, 'Street address is required').transform((s) => s.trim()),
      city: z.string().min(1, 'City is required').transform((s) => s.trim()),
      postalCode: z.string().default(''),
      country: countrySchema
    }),
    recipientContacts: z.object({
      recipientName: z.string().default(''),
      recipientPhone: z.string().min(1, 'Phone number is required for delivery').transform((s) => s.trim()),
      recipientEmail: z
        .string()
        .default('')
        .transform((v) => (v ?? '').trim())
        .refine((v) => v === '' || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v), {
          message: 'Invalid email address'
        })
    }),
    deliveryDate: z.string().min(1, 'Delivery date is required'),
    selectedTimeSlot: z
      .union([z.literal(''), timeSlotLabelSchema])
      .refine((v) => v !== '', { message: 'Delivery time slot is required' }),
    priority: prioritySchema
  })
  .refine(
    (data) => {
      const date = data.deliveryDate ? new Date(data.deliveryDate + 'T12:00:00') : null;
      if (!date || isNaN(date.getTime())) return true;
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      tomorrow.setHours(0, 0, 0, 0);
      const max = new Date();
      max.setDate(max.getDate() + 14);
      max.setHours(23, 59, 59, 999);
      date.setHours(12, 0, 0, 0);
      return date >= tomorrow && date <= max;
    },
    { message: 'Delivery date must be between tomorrow and 14 days from now', path: ['deliveryDate'] }
  )
  .transform((data) => {
    const slot = TIME_SLOTS.find((s) => s.label === data.selectedTimeSlot);
    const startTime = slot
      ? new Date(`${data.deliveryDate}T${slot.start}:00`).toISOString()
      : '';
    const endTime = slot
      ? new Date(`${data.deliveryDate}T${slot.end}:00`).toISOString()
      : '';
    return {
      deliveryAddress: {
        ...data.deliveryAddress,
        postalCode: data.deliveryAddress.postalCode ?? ''
      },
      recipientContacts: {
        recipientName: data.recipientContacts.recipientName ?? '',
        recipientPhone: data.recipientContacts.recipientPhone,
        recipientEmail: data.recipientContacts.recipientEmail ?? ''
      },
      priority: data.priority,
      deliveryPeriod: { startTime, endTime }
    };
  });

export type CheckoutFormData = z.output<typeof checkoutFormInputSchema>;
type CheckoutFormInput = z.input<typeof checkoutFormInputSchema>;

interface CheckoutFormProps {
  onSubmit: (data: CheckoutFormData) => Promise<void>;
  isLoading?: boolean;
  submitError?: string | null;
}

const defaultValues: CheckoutFormInput = {
  deliveryAddress: {
    street: '',
    city: '',
    postalCode: '',
    country: 'Germany'
  },
  recipientContacts: {
    recipientName: '',
    recipientPhone: '',
    recipientEmail: ''
  },
  deliveryDate: '',
  selectedTimeSlot: '',
  priority: 'NORMAL'
};

export default function CheckoutForm({
  onSubmit,
  isLoading = false,
  submitError = null
}: CheckoutFormProps) {
  const { session } = useSession();
  const {
    register,
    handleSubmit: rhfHandleSubmit,
    setValue,
    watch,
    formState: { errors }
  } = useForm<CheckoutFormInput>({
    resolver: zodResolver(checkoutFormInputSchema) as never,
    defaultValues
  });

  const selectedTimeSlot = watch('selectedTimeSlot');
  const selectedDeliveryDate = watch('deliveryDate');
  const recipientName = watch('recipientContacts.recipientName');
  const recipientPhone = watch('recipientContacts.recipientPhone');
  const recipientEmail = watch('recipientContacts.recipientEmail');
  const tomorrowDate = getMinDate();

  useEffect(() => {
    const traits: Record<string, unknown> =
      (session?.identity?.traits as Record<string, unknown> | undefined) ?? {};
    const nameTraits = (traits.name as Record<string, string> | undefined) ?? {};
    const fullName = `${nameTraits.first ?? ''} ${nameTraits.last ?? ''}`.trim();
    const email = typeof traits.email === 'string' ? traits.email : '';
    const phone = typeof traits.phone === 'string' ? traits.phone : '';

    if (!recipientName && fullName) {
      setValue('recipientContacts.recipientName', fullName, { shouldDirty: false });
    }
    if (!recipientPhone && phone) {
      setValue('recipientContacts.recipientPhone', phone, { shouldDirty: false });
    }
    if (!recipientEmail && email) {
      setValue('recipientContacts.recipientEmail', email, { shouldDirty: false });
    }
  }, [recipientEmail, recipientName, recipientPhone, session, setValue]);

  const onValid = async (data: CheckoutFormData) => {
    await onSubmit(data);
  };

  return (
    <form
      onSubmit={rhfHandleSubmit((data) => onValid(data as unknown as CheckoutFormData))}
      className="space-y-6"
    >
      {/* Delivery Address Section */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-[var(--color-foreground)]">Delivery Address</h3>

        <div>
          <label
            htmlFor="street"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Street Address *
          </label>
          <input
            type="text"
            id="street"
            {...register('deliveryAddress.street')}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] bg-[var(--color-surface)] text-[var(--color-foreground)] ${
              errors.deliveryAddress?.street
                ? 'border-[var(--color-destructive)]'
                : 'border-[var(--color-border)]'
            }`}
            placeholder="123 Main Street, Apt 4"
          />
          {errors.deliveryAddress?.street && (
            <p className="mt-1 text-sm text-[var(--color-destructive)]">{errors.deliveryAddress.street.message}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label
              htmlFor="city"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            City *
            </label>
            <input
              type="text"
              id="city"
              {...register('deliveryAddress.city')}
              className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] bg-[var(--color-surface)] text-[var(--color-foreground)] ${
                errors.deliveryAddress?.city
                  ? 'border-[var(--color-destructive)]'
                  : 'border-[var(--color-border)]'
              }`}
              placeholder="Berlin"
            />
            {errors.deliveryAddress?.city && (
              <p className="mt-1 text-sm text-[var(--color-destructive)]">{errors.deliveryAddress.city.message}</p>
            )}
          </div>

          <div>
            <label
              htmlFor="postalCode"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Postal Code
            </label>
            <input
              type="text"
              id="postalCode"
              {...register('deliveryAddress.postalCode')}
              className="mt-1 block w-full rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]"
              placeholder="10115"
            />
          </div>
        </div>

        <div>
          <label
            htmlFor="country"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Country *
          </label>
          <select
            id="country"
            {...register('deliveryAddress.country')}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] bg-[var(--color-surface)] text-[var(--color-foreground)] ${
              errors.deliveryAddress?.country
                ? 'border-[var(--color-destructive)]'
                : 'border-[var(--color-border)]'
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
            <p className="mt-1 text-sm text-[var(--color-destructive)]">{errors.deliveryAddress.country.message}</p>
          )}
        </div>
      </div>

      {/* Recipient Contacts Section */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-[var(--color-foreground)]">Recipient Contacts</h3>

        <div>
          <label
            htmlFor="recipientName"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Recipient Name
          </label>
          <input
            type="text"
            id="recipientName"
            {...register('recipientContacts.recipientName')}
            className="mt-1 block w-full rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]"
            placeholder="John Doe"
          />
        </div>

        <div>
          <label
            htmlFor="recipientPhone"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Phone *
          </label>
          <input
            type="tel"
            id="recipientPhone"
            {...register('recipientContacts.recipientPhone')}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] bg-[var(--color-surface)] text-[var(--color-foreground)] ${
                errors.recipientContacts?.recipientPhone
                  ? 'border-[var(--color-destructive)]'
                  : 'border-[var(--color-border)]'
              }`}
            placeholder="+49 123 456 7890"
          />
          {errors.recipientContacts?.recipientPhone && (
            <p className="mt-1 text-sm text-[var(--color-destructive)]">
              {errors.recipientContacts.recipientPhone.message}
            </p>
          )}
        </div>

        <div>
          <label
            htmlFor="recipientEmail"
            className="block text-sm font-medium text-[var(--color-muted-foreground)]"
          >
            Email
          </label>
          <input
            type="email"
            id="recipientEmail"
            {...register('recipientContacts.recipientEmail')}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] bg-[var(--color-surface)] text-[var(--color-foreground)] ${
                errors.recipientContacts?.recipientEmail
                  ? 'border-[var(--color-destructive)]'
                  : 'border-[var(--color-border)]'
              }`}
            placeholder="recipient@example.com"
          />
          {errors.recipientContacts?.recipientEmail && (
            <p className="mt-1 text-sm text-[var(--color-destructive)]">
              {errors.recipientContacts.recipientEmail.message}
            </p>
          )}
        </div>
      </div>

      {/* Delivery Period Section */}
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
            max={getMaxDate()}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] bg-[var(--color-surface)] text-[var(--color-foreground)] ${
                errors.deliveryDate
                  ? 'border-[var(--color-destructive)]'
                  : 'border-[var(--color-border)]'
            }`}
          />
          {errors.deliveryDate && (
            <p className="mt-1 text-sm text-[var(--color-destructive)]">{errors.deliveryDate.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-[var(--color-muted-foreground)]">
            Time Slot *
          </label>
          <div className="mt-2 grid grid-cols-2 gap-2">
            {TIME_SLOTS.map((slot) => (
              <button
                key={slot.label}
                type="button"
                onClick={() => setValue('selectedTimeSlot', slot.label, { shouldValidate: true })}
                className={`rounded-md border px-4 py-2 text-sm font-medium transition-colors ${
                  selectedTimeSlot === slot.label
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)] text-white'
                    : 'border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)]'
                }`}
              >
                {slot.label}
              </button>
            ))}
          </div>
          {errors.selectedTimeSlot && (
            <p className="mt-1 text-sm text-[var(--color-destructive)]">{errors.selectedTimeSlot.message}</p>
          )}
        </div>
      </div>

      {/* Priority Section */}
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
            <span className="ml-2 text-sm text-[var(--color-muted-foreground)]">
              Normal Delivery
            </span>
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

      {/* Submit Button */}
      {submitError ? (
        <div
          role="alert"
          aria-live="assertive"
          className="rounded-lg border border-[var(--color-border)] bg-[var(--color-muted)] p-3 text-sm text-[var(--color-foreground)]"
        >
          {submitError}
        </div>
      ) : null}
      <Button type="submit" loading={isLoading} className="w-full justify-center">
        Place Order
      </Button>
    </form>
  );
}
