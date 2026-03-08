'use client';

import { AppHeader } from '@shortlink-org/ui-kit';
import { AxiosError } from 'axios';
import CartModal from 'components/cart/modal';
import LogoSquare from 'components/logo-square';
import { useSession } from '@/contexts/SessionContext';
import ory from '@/lib/ory/sdk';
import { createUrl } from 'lib/utils';
import Link from 'next/link';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
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
  const { session, hasSession, isLoading } = useSession();
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [logoutToken, setLogoutToken] = useState('');
  const menu = [
    { name: 'Home', href: '/' },
    { name: 'Shop', href: '/search' }
  ];
  const loginUrl = process.env.NEXT_PUBLIC_LOGIN_URL || '/auth/login';
  const traits = (session?.identity?.traits as Record<string, unknown> | undefined) ?? {};
  const nameTraits = traits.name as Record<string, string> | undefined;
  const firstName = nameTraits?.first ?? '';
  const lastName = nameTraits?.last ?? '';
  const email = (traits.email as string | undefined) ?? '';
  const displayName = `${firstName} ${lastName}`.trim() || email || 'User';

  useEffect(() => {
    if (!hasSession) {
      setLogoutToken('');
      return;
    }

    ory
      .createBrowserLogoutFlow()
      .then(({ data }) => {
        setLogoutToken(data.logout_token);
      })
      .catch((err: AxiosError) => {
        if (err.response?.status === 401) return;
        return Promise.reject(err);
      });
  }, [hasSession]);

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
    <div className="relative border-b border-[var(--color-border)] bg-[linear-gradient(180deg,var(--color-background)_0%,color-mix(in_srgb,var(--color-surface)_78%,transparent)_100%)]">
      <AppHeader
        className="mx-auto max-w-7xl px-3 pt-3 sm:px-4 lg:px-6"
        brand={{
          name: SITE_NAME || 'Shortlink Shop',
          href: '/',
          render: ({ href, className }) => (
            <Link
              href={href}
              prefetch={true}
              className={className}
            >
              <span className="inline-flex items-center gap-3 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-2 shadow-[0_18px_44px_-36px_rgba(15,23,42,0.26)]">
                <LogoSquare size="sm" />
                <span className="min-w-0">
                  <span className="block text-sm font-semibold tracking-tight text-[var(--color-foreground)]">
                    {SITE_NAME || 'Shortlink Shop'}
                  </span>
                  <span className="block text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--color-muted-foreground)]">
                    Curated storefront
                  </span>
                </span>
              </span>
            </Link>
          )
        }}
        workspaceLabel="Commerce"
        statusBadge={{ label: 'Live', tone: 'accent' }}
        navigation={menu}
        currentPath={pathname}
        LinkComponent={HeaderLink}
        showMenuButton={false}
        showThemeToggle={true}
        showSearch={true}
        showNotifications={true}
        notifications={{
          render: () => <CartModal />
        }}
        searchProps={{
          placeholder: 'Search for products...',
          defaultQuery: searchParams?.get('q') || '',
          onSearch: handleSearch
        }}
        showProfile={hasSession && !isLoading}
        profile={
          hasSession
            ? {
                name: displayName,
                email,
                menuItems: [
                  { name: 'Your Profile', href: '/profile' },
                  { name: 'Orders', href: '/orders' },
                  {
                    name: 'Sign out',
                    onClick: async () => {
                      if (!logoutToken) return;
                      await ory.updateLogoutFlow({ token: logoutToken });
                      window.location.assign(loginUrl);
                    },
                    confirmDialog: {
                      title: 'Sign out?',
                      description: 'You will need to log in again to access your account.',
                      confirmText: 'Sign out',
                      variant: 'danger'
                    }
                  }
                ]
              }
            : undefined
        }
        showLogin={!hasSession && !isLoading}
        loginButton={{ href: loginUrl, label: 'Sign in' }}
        sticky={false}
        fullWidth={true}
      />
      <div className="mx-auto max-w-7xl px-3 pb-4 pt-2 sm:px-4 lg:px-6 md:hidden">
        <Suspense fallback={<SearchSkeleton />}>
          <Search
            className="max-w-none"
            inputClassName="rounded-full border-[var(--color-border)] bg-[var(--color-surface)] px-5 py-3 text-[var(--color-foreground)] shadow-[0_16px_36px_-32px_rgba(15,23,42,0.24)] placeholder:text-[var(--color-muted-foreground)]"
          />
        </Suspense>
      </div>
    </div>
  );
}
