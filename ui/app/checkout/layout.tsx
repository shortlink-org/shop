import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Checkout',
  description: 'Complete your order delivery details and place your order.'
};

export default function CheckoutLayout({
  children
}: {
  children: React.ReactNode;
}) {
  return children;
}
