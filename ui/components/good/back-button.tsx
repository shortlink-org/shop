'use client';

import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';

type BackButtonProps = {
  goodId: number;
  className?: string;
};

/**
 * When viewing a specific image (?image=N), "Back" goes to the product page (no query).
 * Otherwise links to search (listing).
 */
export function BackButton({ goodId, className = '' }: BackButtonProps) {
  const searchParams = useSearchParams();
  const imageParam = searchParams.get('image');

  const baseClassName =
    'inline-flex items-center gap-2 text-sm text-neutral-500 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100 transition-colors';

  if (imageParam != null) {
    return (
      <Link
        href={`/good/${goodId}`}
        className={`${baseClassName} ${className}`}
        aria-label="Back to product"
      >
        <ArrowLeftIcon className="h-4 w-4" />
        <span>Back</span>
      </Link>
    );
  }

  return (
    <Link
      href="/search"
      className={`${baseClassName} ${className}`}
      aria-label="Back to search"
    >
      <ArrowLeftIcon className="h-4 w-4" />
      <span>Back</span>
    </Link>
  );
}
