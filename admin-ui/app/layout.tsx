import type { Metadata } from 'next';
import { AdminShell } from '@/components/admin/admin-shell';
import { AdminProviders } from '@/providers/admin-providers';
import '@shortlink-org/ui-kit/dist/assets/index.css';
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
    <html lang="en" suppressHydrationWarning>
      <body>
        <AdminProviders>
          <AdminShell>{children}</AdminShell>
        </AdminProviders>
      </body>
    </html>
  );
}
