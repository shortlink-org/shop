import type { Metadata } from 'next';
import { Suspense } from 'react';
import { OrdersPageClient } from './orders-page-client';

export const metadata: Metadata = {
  title: 'Order lookup',
  description: 'Search by order ID to inspect order state, delivery window, and tracking.'
};

export default function OrdersPage() {
  return (
    <Suspense fallback={<div className="admin-card p-8 animate-pulse">Loading...</div>}>
      <OrdersPageClient />
    </Suspense>
  );
}
