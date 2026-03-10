'use client';

import { Sidebar, type SidebarSection } from '@shortlink-org/ui-kit';
import {
  ClipboardDocumentListIcon,
  FireIcon,
  HomeIcon,
  RocketLaunchIcon,
  ShoppingBagIcon,
  SparklesIcon
} from '@heroicons/react/24/outline';
import { SORT_SLUGS } from 'lib/constants';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const sections: SidebarSection[] = [
  {
    type: 'simple',
    items: [
      {
        url: '/',
        icon: <HomeIcon />,
        name: 'Home'
      },
      {
        url: '/search',
        icon: <ShoppingBagIcon />,
        name: 'Catalog'
      },
      {
        url: '/checkout',
        icon: <ClipboardDocumentListIcon />,
        name: 'Checkout'
      }
    ]
  },
  {
    type: 'collapsible',
    icon: SparklesIcon,
    title: 'Discover',
    items: [
      {
        url: `/search?sort=${SORT_SLUGS.latest}`,
        icon: <RocketLaunchIcon />,
        name: 'New arrivals'
      },
      {
        url: `/search?sort=${SORT_SLUGS.trending}`,
        icon: <FireIcon />,
        name: 'Best sellers'
      }
    ]
  }
];

export function ShopSidebar() {
  const pathname = usePathname();

  return (
    <Sidebar
      sections={sections}
      activePath={pathname}
      variant="sticky"
      height="calc(100vh - 6.5rem)"
      className="shop-sidebar top-6"
      footerSlot={
        <div className="border-t border-[var(--color-border)] p-3">
          <div className="rounded-[1.2rem] border border-[var(--color-border)] bg-[var(--color-background)]/75 p-4">
            <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--color-muted-foreground)]">
              Marketplace
            </p>
            <p className="mt-2 text-sm font-semibold text-[var(--color-foreground)]">
              Track storefront momentum
            </p>
            <p className="mt-2 text-sm leading-6 text-[var(--color-muted-foreground)]">
              Follow this week&apos;s fastest movers and browse what is converting now.
            </p>
            <Link
              href="/#leaderboard"
              className="mt-4 inline-flex items-center rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] text-[var(--color-foreground)] transition-colors hover:bg-[var(--color-muted)]"
            >
              View leaderboard
            </Link>
          </div>
        </div>
      }
    />
  );
}
