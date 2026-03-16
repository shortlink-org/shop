import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Register courier',
  description: 'Register a new courier in the delivery system.'
};

export default function CreateCourierLayout({ children }: { children: React.ReactNode }) {
  return children;
}
