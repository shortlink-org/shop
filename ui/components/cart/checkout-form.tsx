'use client';

'use no memo'; // react-hook-form's setValue in useCallback is incompatible with React Compiler

import { Button } from '@/lib/ui-kit';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useSession } from '@/contexts/SessionContext';
import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';
import { fetchRandomAddress as fetchRandomAddressAction } from './actions';
import { DeliveryAddressSection } from './checkout-form/delivery-address-section';
import { DeliveryPeriodSection } from './checkout-form/delivery-period-section';
import { PrioritySection } from './checkout-form/priority-section';
import { RecipientContactsSection } from './checkout-form/recipient-contacts-section';

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

const MIN_DELIVERY_OFFSET_DAYS = 1;
const MAX_DELIVERY_OFFSET_DAYS = 14;

type CalendarDate = {
  year: number;
  month: number;
  day: number;
};

function formatCalendarDate(value: CalendarDate): string {
  const month = String(value.month).padStart(2, '0');
  const day = String(value.day).padStart(2, '0');
  return `${value.year}-${month}-${day}`;
}

function getToday(): CalendarDate {
  const now = new Date();
  return {
    year: now.getFullYear(),
    month: now.getMonth() + 1,
    day: now.getDate()
  };
}

function addDays(value: CalendarDate, days: number): CalendarDate {
  const next = new Date(value.year, value.month - 1, value.day);
  next.setDate(next.getDate() + days);
  return {
    year: next.getFullYear(),
    month: next.getMonth() + 1,
    day: next.getDate()
  };
}

function compareCalendarDates(left: CalendarDate, right: CalendarDate): number {
  if (left.year !== right.year) return left.year - right.year;
  if (left.month !== right.month) return left.month - right.month;
  return left.day - right.day;
}

function getMinDate(): string {
  return formatCalendarDate(addDays(getToday(), MIN_DELIVERY_OFFSET_DAYS));
}

function getMaxDate(): string {
  return formatCalendarDate(addDays(getToday(), MAX_DELIVERY_OFFSET_DAYS));
}

function parseDeliveryDate(value: string): CalendarDate | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match) {
    return null;
  }

  const [, yearRaw, monthRaw, dayRaw] = match;
  const year = Number(yearRaw);
  const month = Number(monthRaw);
  const day = Number(dayRaw);
  const date = new Date(year, month - 1, day);

  if (
    Number.isNaN(date.getTime()) ||
    date.getFullYear() !== year ||
    date.getMonth() + 1 !== month ||
    date.getDate() !== day
  ) {
    return null;
  }

  return { year, month, day };
}

function createDeliveryTimestamp(deliveryDate: string, time: string): string {
  const parsedDate = parseDeliveryDate(deliveryDate);
  const timeMatch = /^(\d{2}):(\d{2})$/.exec(time);

  if (!parsedDate || !timeMatch) {
    return '';
  }

  const [, hourRaw, minuteRaw] = timeMatch;
  const hour = Number(hourRaw);
  const minute = Number(minuteRaw);
  const localDateTime = new Date(
    parsedDate.year,
    parsedDate.month - 1,
    parsedDate.day,
    hour,
    minute,
    0,
    0
  );

  return localDateTime.toISOString();
}

const countrySchema = z.enum(COUNTRY_VALUES);
const prioritySchema = z.enum(['NORMAL', 'URGENT']);
const timeSlotLabelSchema = z.enum(TIME_SLOT_LABELS as unknown as [string, ...string[]]);

const checkoutFormInputSchema = z
  .object({
    deliveryAddress: z.object({
      street: z
        .string()
        .min(1, 'Street address is required')
        .transform((s) => s.trim()),
      city: z
        .string()
        .min(1, 'City is required')
        .transform((s) => s.trim()),
      country: countrySchema
    }),
    recipientContacts: z.object({
      recipientName: z.string().default(''),
      recipientPhone: z
        .string()
        .min(1, 'Phone number is required for delivery')
        .transform((s) => s.trim()),
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
      const date = parseDeliveryDate(data.deliveryDate);
      if (!date) return false;

      const today = getToday();
      const tomorrow = addDays(today, MIN_DELIVERY_OFFSET_DAYS);
      const max = addDays(today, MAX_DELIVERY_OFFSET_DAYS);

      return compareCalendarDates(date, tomorrow) >= 0 && compareCalendarDates(date, max) <= 0;
    },
    {
      message: 'Delivery date must be between tomorrow and 14 days from now',
      path: ['deliveryDate']
    }
  )
  .transform((data) => {
    const slot = TIME_SLOTS.find((s) => s.label === data.selectedTimeSlot);
    const startTime = slot ? createDeliveryTimestamp(data.deliveryDate, slot.start) : '';
    const endTime = slot ? createDeliveryTimestamp(data.deliveryDate, slot.end) : '';
    return {
      deliveryAddress: data.deliveryAddress,
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
export type CheckoutFormInput = z.input<typeof checkoutFormInputSchema>;

interface CheckoutFormProps {
  onSubmit: (data: CheckoutFormData) => Promise<void>;
  isLoading?: boolean;
  submitError?: string | null;
}

const defaultValues: CheckoutFormInput = {
  deliveryAddress: {
    street: '',
    city: '',
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
  const sessionDefaults = useMemo(() => {
    const traits: Record<string, unknown> =
      (session?.identity?.traits as Record<string, unknown> | undefined) ?? {};
    const nameTraits = (traits.name as Record<string, string> | undefined) ?? {};
    const fullName = `${nameTraits.first ?? ''} ${nameTraits.last ?? ''}`.trim();
    const email = typeof traits.email === 'string' ? traits.email : '';
    const phone = typeof traits.phone === 'string' ? traits.phone : '';
    return {
      ...defaultValues,
      recipientContacts: {
        ...defaultValues.recipientContacts,
        recipientName: fullName || defaultValues.recipientContacts.recipientName,
        recipientPhone: phone || defaultValues.recipientContacts.recipientPhone,
        recipientEmail: email || defaultValues.recipientContacts.recipientEmail
      }
    };
  }, [session]);

  const {
    register,
    handleSubmit: rhfHandleSubmit,
    setValue,
    watch,
    formState: { errors }
  } = useForm<CheckoutFormInput>({
    resolver: zodResolver(checkoutFormInputSchema) as never,
    defaultValues: sessionDefaults
  });

  // eslint-disable-next-line react-hooks/incompatible-library -- react-hook-form watch() not memoizable; acceptable for form UI
  const selectedTimeSlot = watch('selectedTimeSlot');
  const selectedDeliveryDate = watch('deliveryDate');
  const tomorrowDate = getMinDate();
  const maxDeliveryDate = getMaxDate();
  const [randomAddressLoading, setRandomAddressLoading] = useState(false);

  const handleRandomAddress = useCallback(async () => {
    setRandomAddressLoading(true);
    try {
      const result = await fetchRandomAddressAction();
      if (!result.ok) {
        toast.error(result.message);
        setRandomAddressLoading(false);
        return;
      }

      const addr = result.address;
      setValue('deliveryAddress.street', addr.street ?? '', {
        shouldDirty: true,
        shouldTouch: true
      });
      setValue('deliveryAddress.city', addr.city ?? '', { shouldDirty: true, shouldTouch: true });
      setValue(
        'deliveryAddress.country',
        (addr.country ?? 'Germany') as CheckoutFormInput['deliveryAddress']['country'],
        {
          shouldDirty: true,
          shouldTouch: true
        }
      );
    } catch {
      // Ignore - user can retry
    }
    setRandomAddressLoading(false);
  }, [setValue]);

  const onValid = async (data: CheckoutFormData) => {
    await onSubmit(data);
  };

  return (
    <form
      onSubmit={rhfHandleSubmit((data) => onValid(data as unknown as CheckoutFormData))}
      className="space-y-6"
    >
      <DeliveryAddressSection
        register={register}
        errors={errors}
        onRandomAddress={handleRandomAddress}
        randomAddressLoading={randomAddressLoading}
      />

      <RecipientContactsSection register={register} errors={errors} />

      <DeliveryPeriodSection
        register={register}
        setValue={setValue}
        selectedTimeSlot={selectedTimeSlot}
        selectedDeliveryDate={selectedDeliveryDate}
        errors={errors}
        tomorrowDate={tomorrowDate}
        maxDeliveryDate={maxDeliveryDate}
      />

      <PrioritySection register={register} />

      {/* Submit Button */}
      {submitError ? (
        <div
          role="alert"
          aria-live="assertive"
          className="rounded-lg border border-red-300 bg-red-50 p-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-950/50 dark:text-red-200"
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
