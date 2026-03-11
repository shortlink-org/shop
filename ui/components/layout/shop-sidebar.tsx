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
      className="shop-sidebar top-8"
      footerSlot={
        <div className="border-t border-[var(--color-border)]/80 p-3">
          <div className="rounded-[1.35rem] border border-[var(--color-border)] bg-[linear-gradient(180deg,color-mix(in_srgb,var(--color-surface)_94%,transparent),color-mix(in_srgb,var(--color-muted)_76%,transparent))] p-4 shadow-[0_18px_44px_-36px_rgba(15,23,42,0.28)]">
            <p className="text-[11px] font-semibold tracking-[0.18em] text-[var(--color-muted-foreground)] uppercase">
              Campaign radar
            </p>
            <p className="mt-2 text-sm font-semibold text-[var(--color-foreground)]">
              Follow what is moving now
            </p>
            <p className="mt-2 text-sm leading-6 text-[var(--color-muted-foreground)]">
              Jump from the storefront rail straight into this week&apos;s highest-grossing and
              fastest-selling goods.
            </p>
            <Link
              href="/#leaderboard"
              className="mt-4 inline-flex items-center rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-xs font-semibold tracking-[0.16em] text-[var(--color-foreground)] uppercase transition-colors hover:bg-[var(--color-muted)]"
            >
              View leaderboard
            </Link>
          </div>
        </div>
      }
    />
  );
}
