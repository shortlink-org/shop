import type { Metadata } from 'next';
import { Suspense } from 'react';
import { RefineProvider } from '@/providers/refine-provider';
import './globals.css';

export const metadata: Metadata = {
  title: 'Delivery Admin',
  description: 'Admin panel for Delivery service management',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ru">
      <body>
        <Suspense fallback={<div>Loading...</div>}>
          <RefineProvider>{children}</RefineProvider>
        </Suspense>
      </body>
    </html>
  );
}
