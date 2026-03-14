'use client';

import { Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';

export default function OrdersPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialOrderId = searchParams.get('orderId') ?? '';
  const [orderId, setOrderId] = useState(() => initialOrderId);

  useEffect(() => {
    if (initialOrderId) {
      router.replace(`/orders/${initialOrderId}`);
    }
  }, [initialOrderId, router]);

  const submitLookup = () => {
    const normalized = orderId.trim();
    if (!normalized) return;
    router.push(`/orders/${normalized}`);
  };

  return (
    <div className="space-y-6">
      <section className="admin-card overflow-hidden p-6 sm:p-8">
        <div className="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
          <div className="space-y-3">
            <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
              Order operations
            </p>
            <div>
              <h1 className="text-3xl font-semibold tracking-tight">Order lookup</h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--color-muted-foreground)]">
                Search by `orderId` to inspect order state, delivery window, assigned courier, and live
                tracking snapshot from the BFF.
              </p>
            </div>
          </div>

          <div className="admin-card flex w-full max-w-xl flex-col gap-3 p-4">
            <label className="admin-field">
              <span className="admin-label">Order ID</span>
              <input
                className="admin-input font-mono"
                placeholder="e.g. order-123"
                value={orderId}
                onChange={(event) => setOrderId(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.preventDefault();
                    submitLookup();
                  }
                }}
              />
            </label>
            <div className="flex flex-wrap gap-3">
              <Button onClick={submitLookup}>Lookup order</Button>
            </div>
          </div>
        </div>
      </section>

      <div className="py-8">
        <FeedbackPanel
          variant="empty"
          eyebrow="Order operations"
          title="Enter an order ID to begin"
          message="This route is now the lookup entry point. Detailed inspection lives on the dedicated `/orders/[id]` page."
          className="mx-auto max-w-3xl"
        />
      </div>
    </div>
  );
}
