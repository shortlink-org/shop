'use client';

import { AppHeader } from '@/lib/ui-kit';
import { AxiosError } from 'axios';
import CartModal from 'components/cart/modal';
import LogoSquare from 'components/logo-square';
import { useSession } from '@/contexts/SessionContext';
import ory from '@/lib/ory/sdk';
import Link from 'next/link';
import { createUrl } from 'lib/utils';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useCallback, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { Suspense } from 'react';

const { SITE_NAME } = process.env;
const headerSlotClassNames = {
  container: 'mx-auto max-w-7xl px-3 sm:px-4 lg:px-6',
  header: 'min-h-[4.5rem] gap-3 py-3 md:min-h-[5rem]',
  brandRail: 'min-w-0 shrink-0 gap-3',
  // Kill default absolute positioning on small viewports (overlaps brand / creates stacked “ghost” pills).
  controlsRail:
    '!static !inset-auto !right-auto z-10 ml-auto flex shrink-0 flex-nowrap items-center justify-end gap-3 pr-2 sm:ml-6 sm:pr-0 xl:ml-0 xl:justify-self-end',
  themeToggle: 'relative mr-1 hidden shrink-0 overflow-visible xl:flex xl:items-center',
  navigation: 'hidden',
  search: 'min-w-0 max-md:max-w-[min(18rem,calc(100vw-10rem))]',
  notifications: 'ml-0',
  notificationsButton: 'min-h-[2.75rem] px-0 py-2.5',
  profile: 'ml-0',
  profileButton: 'min-h-[2.75rem]',
  loginButton: 'min-h-[2.75rem] py-2.5'
};

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

function NavbarContent() {
  const { session, hasSession, isLoading } = useSession();
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [logoutToken, setLogoutToken] = useState('');
  const loginUrl = process.env.NEXT_PUBLIC_LOGIN_URL || '/auth/login';

  const defaultQuery = searchParams.get('q') ?? '';

  const onHeaderSearch = useCallback(
    (query: string) => {
      const trimmed = query.trim();
      const next = new URLSearchParams(searchParams.toString());
      if (trimmed) {
        next.set('q', trimmed);
      } else {
        next.delete('q');
      }
      router.push(createUrl('/search', next));
    },
    [router, searchParams]
  );
  const traits = (session?.identity?.traits as Record<string, unknown> | undefined) ?? {};
  const nameTraits = traits.name as Record<string, string> | undefined;
  const firstName = nameTraits?.first ?? '';
  const lastName = nameTraits?.last ?? '';
  const email = (traits.email as string | undefined) ?? '';
  const displayName = `${firstName} ${lastName}`.trim() || email || 'User';

  useEffect(() => {
    if (!hasSession) {
      queueMicrotask(() => setLogoutToken(''));
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

  return (
    <div className="shop-navbar relative border-b border-[var(--color-border)] bg-[linear-gradient(180deg,var(--color-background)_0%,color-mix(in_srgb,var(--color-surface)_78%,transparent)_100%)]">
      <AppHeader
        className="shop-navbar__header"
        brand={{
          name: SITE_NAME || 'Shortlink Shop',
          href: '/',
          render: ({ href, className }) => (
            <Link href={href} prefetch={true} className={`${className ?? ''} min-w-0`}>
              <span className="inline-flex min-w-0 items-center gap-2.5 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] py-2 pr-4 pl-3 shadow-[0_18px_44px_-36px_rgba(15,23,42,0.26)]">
                <LogoSquare size="sm" />
                <span className="min-w-0 max-w-[12rem] sm:max-w-none">
                  <span className="block text-sm font-semibold tracking-tight text-[var(--color-foreground)]">
                    {SITE_NAME || 'Shortlink Shop'}
                  </span>
                  <span className="hidden text-[11px] font-semibold tracking-[0.18em] text-[var(--color-muted-foreground)] uppercase sm:block">
                    Curated storefront
                  </span>
                </span>
              </span>
            </Link>
          )
        }}
        navigation={[]}
        currentPath={pathname}
        LinkComponent={HeaderLink}
        showMenuButton={false}
        showThemeToggle={true}
        showSearch={true}
        searchProps={{
          placeholder: 'Search for products…',
          defaultQuery,
          onSearch: onHeaderSearch
        }}
        showNotifications={true}
        notifications={{
          render: () => <CartModal />
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
        slotClassNames={headerSlotClassNames}
      />
    </div>
  );
}

export function Navbar() {
  return (
    <Suspense fallback={<div className="min-h-[5.5rem] border-b border-[var(--color-border)]" />}>
      <NavbarContent />
    </Suspense>
  );
}
