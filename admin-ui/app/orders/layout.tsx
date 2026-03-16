import type { Metadata } from 'next';
import { Suspense } from 'react';

export const metadata: Metadata = {
  title: 'Order operations',
  description: 'Order lookup and delivery tracking management.'
};

export default function OrdersLayout({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<div className="admin-card p-8 animate-pulse">Loading...</div>}>{children}</Suspense>;
}
