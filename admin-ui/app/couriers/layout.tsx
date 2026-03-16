import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Couriers',
  description: 'Courier management and workspace.'
};

export default function CouriersLayout({ children }: { children: React.ReactNode }) {
  return children;
}
