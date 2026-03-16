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

const socialLinks = [
  {
    name: 'GitHub',
    href: 'https://github.com/shortlink-org/shortlink',
    iconPath:
      'M896 128q209 0 385.5 103t279.5 279.5 103 385.5q0 251-146.5 451.5t-378.5 277.5q-27 5-40-7t-13-30q0-3 .5-76.5t.5-134.5q0-97-52-142 57-6 102.5-18t94-39 81-66.5 53-105 20.5-150.5q0-119-79-206 37-91-8-204-28-9-81 11t-92 44l-38 24q-93-26-192-26t-192 26q-16-11-42.5-27t-83.5-38.5-85-13.5q-45 113-8 204-79 87-79 206 0 85 20.5 150t52.5 105 80.5 67 94 39 102.5 18q-39 36-49 103-21 10-45 15t-57 5-65.5-21.5-55.5-62.5q-19-32-48.5-52t-49.5-24l-20-3q-21 0-29 4.5t-5 11.5 9 14 13 12l7 5q22 10 43.5 38t31.5 51l10 23q13 38 44 61.5t67 30 69.5 7 55.5-3.5l23-4q0 38 .5 88.5t.5 54.5q0 18-13 30t-40 7q-232-77-378.5-277.5t-146.5-451.5q0-209 103-385.5t279.5-279.5 385.5-103z'
  },
  {
    name: 'LinkedIn',
    href: 'https://linkedin.com/company/shortlink',
    iconPath:
      'M477 625v991h-330v-991h330zm21-306q1 73-50.5 122t-135.5 49h-2q-82 0-132-49t-50-122q0-74 51.5-122.5t134.5-48.5 133 48.5 51 122.5zm1166 729v568h-329v-530q0-105-40.5-164.5t-126.5-59.5q-63 0-105.5 34.5t-63.5 85.5q-11 30-11 81v553h-329q2-399 2-647t-1-296l-1-48h329v144h-2q20-32 41-56t56.5-52 87-43.5 114.5-15.5q171 0 275 113.5t104 332.5z'
  }
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
    <div className="border-t border-[var(--color-border)] bg-[linear-gradient(180deg,var(--color-background)_0%,color-mix(in_srgb,var(--color-surface)_82%,transparent)_100%)] pt-10 pb-8">
      <UiKitFooter
        className="w-full border-x-0 border-b-0 border-[var(--color-border)]/80 bg-[color-mix(in_srgb,var(--color-surface)_88%,transparent)] shadow-none backdrop-blur-xl"
        contained={false}
        rounded={false}
        withTopMargin={false}
        links={footerLinks}
        socialLinks={socialLinks}
        LinkComponent={FooterLink}
        logoSlot={
          <div className="inline-flex items-center gap-3 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-2">
            <span className="inline-flex size-9 items-center justify-center rounded-full bg-sky-500 text-sm font-semibold text-white">
              SL
            </span>
            <span className="text-sm font-semibold text-[var(--color-foreground)]">
              {SITE_NAME}
            </span>
          </div>
        }
        description="Curated everyday picks for a cleaner, faster storefront experience."
        copyright={
          <span className="text-sm text-[var(--color-muted-foreground)]">
            &copy; {new Date().getFullYear()} {SITE_NAME}. Built with ShortLink Commerce.
          </span>
        }
      />
    </div>
  );
}
