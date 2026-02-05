'use client';

import { useRouter } from 'next/navigation';

export function RetryButton({ children = 'Retry' }: { children?: React.ReactNode }) {
  const router = useRouter();
  return (
    <button
      type="button"
      className="mx-auto mt-4 flex items-center justify-center rounded-full bg-blue-600 px-6 py-3 tracking-wide text-white hover:opacity-90"
      onClick={() => router.refresh()}
    >
      {children}
    </button>
  );
}
