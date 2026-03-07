'use client';

import { Footer as UiKitFooter } from '@shortlink-org/ui-kit';
import { SORT_SLUGS } from 'lib/constants';
import Link from 'next/link';
import type { ReactNode } from 'react';

const SITE_NAME = process.env.NEXT_PUBLIC_SITE_NAME || 'Shop';

const footerLinks = [
  { id: 'all-products', label: 'All Products', href: '/search' },
  { id: 'new-arrivals', label: 'New Arrivals', href: `/search?sort=${SORT_SLUGS.latest}` },
  { id: 'best-sellers', label: 'Best Sellers', href: `/search?sort=${SORT_SLUGS.trending}` }
];

function FooterLink({
  href,
  className,
  target,
  children
}: {
  href: string;
  className?: string;
  target?: string;
  children: ReactNode;
}) {
  return (
    <Link href={href} prefetch={true} className={className} target={target}>
      {children}
    </Link>
  );
}

export function Footer() {
  return (
    <UiKitFooter
      className="mt-0 max-w-none rounded-none border-t border-[var(--color-border)] bg-[var(--color-surface)] px-4 sm:px-6 lg:px-8"
      links={footerLinks}
      socialLinks={[]}
      LinkComponent={FooterLink}
      logoSlot={
        <div className="flex flex-col items-center gap-2">
          <span className="text-xl font-bold tracking-tight text-[var(--color-foreground)]">
            {SITE_NAME}
          </span>
          <span className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--color-muted-foreground)]">
            Curated everyday picks
          </span>
        </div>
      }
      description="Your trusted online shop for quality products."
      copyright={
        <span className="text-sm text-[var(--color-muted-foreground)]">
          &copy; {new Date().getFullYear()} {SITE_NAME}. All rights reserved.
        </span>
      }
    />
  );
}
