'use client';

import { useRouter } from 'next/navigation';

export function RetryButton({ children = 'Retry' }: { children?: React.ReactNode }) {
  const router = useRouter();
  return (
    <button
      type="button"
      className="mx-auto mt-4 flex cursor-pointer items-center justify-center rounded-full border border-[var(--color-border)] bg-[var(--color-foreground)] px-6 py-3 font-medium tracking-wide text-[var(--color-background)] shadow-sm transition-opacity hover:opacity-90"
      onClick={() => router.refresh()}
    >
      {children}
    </button>
  );
}
