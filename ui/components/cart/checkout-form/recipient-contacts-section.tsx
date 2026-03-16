'use client';

import type { UseFormRegister, FieldErrors } from 'react-hook-form';
import type { CheckoutFormInput } from '../checkout-form';

type RecipientContactsSectionProps = {
  register: UseFormRegister<CheckoutFormInput>;
  errors: FieldErrors<CheckoutFormInput>;
};

export function RecipientContactsSection({ register, errors }: RecipientContactsSectionProps) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-[var(--color-foreground)]">Recipient Contacts</h3>

      <div>
        <label htmlFor="recipientName" className="block text-sm font-medium text-[var(--color-muted-foreground)]">
          Recipient Name
        </label>
        <input
          type="text"
          id="recipientName"
          {...register('recipientContacts.recipientName')}
          className="mt-1 block w-full rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none"
          placeholder="John Doe"
        />
      </div>

      <div>
        <label htmlFor="recipientPhone" className="block text-sm font-medium text-[var(--color-muted-foreground)]">
          Phone *
        </label>
        <input
          type="tel"
          id="recipientPhone"
          {...register('recipientContacts.recipientPhone')}
          className={`mt-1 block w-full rounded-md border bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none ${
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
        <label htmlFor="recipientEmail" className="block text-sm font-medium text-[var(--color-muted-foreground)]">
          Email
        </label>
        <input
          type="email"
          id="recipientEmail"
          {...register('recipientContacts.recipientEmail')}
          className={`mt-1 block w-full rounded-md border bg-[var(--color-surface)] px-3 py-2 text-[var(--color-foreground)] shadow-sm focus:ring-2 focus:ring-[var(--color-primary)] focus:outline-none ${
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
  );
}
