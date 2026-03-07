'use client';

import { AppHeader } from '@shortlink-org/ui-kit';
import CartModal from 'components/cart/modal';
import LogoSquare from 'components/logo-square';
import { ProfileWidget } from 'components/user';
import { createUrl } from 'lib/utils';
import Link from 'next/link';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import type { ReactNode } from 'react';
import { Suspense } from 'react';
import Search, { SearchSkeleton } from './search';

const { SITE_NAME } = process.env;

function HeaderLink({
  href,
  className,
  children
}: {
  href: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <Link href={href} prefetch={true} className={className}>
      {children}
    </Link>
  );
}

export function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const menu = [{ name: 'Home', href: '/' }];

  const handleSearch = (query: string) => {
    const newParams = new URLSearchParams(searchParams?.toString() ?? '');

    if (query) {
      newParams.set('q', query);
    } else {
      newParams.delete('q');
    }

    router.push(createUrl('/search', newParams));
  };

  return (
    <div className="relative">
      <AppHeader
        className="border-b-0"
        brand={{
          name: SITE_NAME || 'Shortlink Shop',
          href: '/',
          logo: <LogoSquare size="sm" />
        }}
        navigation={menu}
        currentPath={pathname}
        LinkComponent={HeaderLink}
        showMenuButton={false}
        showThemeToggle={true}
        showSearch={true}
        searchProps={{
          placeholder: 'Search for products...',
          defaultQuery: searchParams?.get('q') || '',
          onSearch: handleSearch
        }}
        showProfile={true}
        profile={{
          render: () => (
            <div className="relative z-30 flex items-center gap-3">
              <ProfileWidget />
              <CartModal />
            </div>
          )
        }}
        sticky={false}
        fullWidth={true}
      />
      <div className="border-b border-[var(--color-border)]/80 bg-[color-mix(in_srgb,var(--color-background)_82%,white_18%)] px-3 pb-4 pt-1 shadow-[0_18px_45px_-40px_rgba(15,23,42,0.24)] backdrop-blur-xl md:hidden">
        <Suspense fallback={<SearchSkeleton />}>
          <Search
            className="max-w-none"
            inputClassName="rounded-2xl border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-3 text-[var(--color-foreground)] shadow-[0_16px_36px_-32px_rgba(15,23,42,0.28)] placeholder:text-[var(--color-muted-foreground)]"
          />
        </Suspense>
      </div>
    </div>
  );
}
