'use client';

import { AppHeader } from '@shortlink-org/ui-kit';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';

import { useSession } from '@/contexts/SessionContext';
import { getLoginUrl, getUserEmail, getUserName } from '@/lib/ory/api';

type NavItem = {
  href: string;
  label: string;
  description: string;
  external?: boolean;
};

const primaryNav: NavItem[] = [
  {
    href: '/',
    label: 'Dashboard',
    description: 'Operations overview',
  },
  {
    href: '/orders',
    label: 'Orders',
    description: 'Lookup and tracking',
  },
  {
    href: '/couriers',
    label: 'Couriers',
    description: 'Fleet and workload',
  },
];

const secondaryNav: NavItem[] = [
  {
    href: 'https://admin.shop.shortlink.best/admin',
    label: 'Django Admin',
    description: 'Model-level backoffice',
    external: true,
  },
];

function HeaderLink({
  href,
  className,
  children,
}: {
  href: string;
  className?: string;
  children: ReactNode;
}) {
  const isExternal = href.startsWith('http://') || href.startsWith('https://');
  if (isExternal) {
    return (
      <a href={href} className={className}>
        {children}
      </a>
    );
  }

  return (
    <Link href={href} className={className}>
      {children}
    </Link>
  );
}

function NavLink({ item, active }: { item: NavItem; active: boolean }) {
  const className = [
    'group flex items-start gap-3 rounded-2xl border px-4 py-3 transition',
    active
      ? 'border-[var(--color-accent)] bg-[color-mix(in_srgb,var(--color-accent)_15%,var(--color-surface))] shadow-[0_16px_36px_-28px_rgba(14,165,233,0.65)]'
      : 'border-transparent bg-transparent hover:border-[var(--color-border)] hover:bg-[color-mix(in_srgb,var(--color-surface)_72%,transparent)]',
  ].join(' ');

  const content = (
    <>
      <span
        className={[
          'mt-1 h-2.5 w-2.5 rounded-full transition',
          active ? 'bg-[var(--color-accent)]' : 'bg-[var(--color-muted-foreground)]/40 group-hover:bg-[var(--color-accent)]/70',
        ].join(' ')}
      />
      <span className="min-w-0">
        <span className="block text-sm font-semibold text-[var(--color-foreground)]">{item.label}</span>
        <span className="block text-xs text-[var(--color-muted-foreground)]">{item.description}</span>
      </span>
    </>
  );

  if (item.external) {
    return (
      <a href={item.href} className={className} target="_blank" rel="noreferrer">
        {content}
      </a>
    );
  }

  return (
    <Link href={item.href} className={className}>
      {content}
    </Link>
  );
}

export function AdminShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const { session, hasSession, isLoading, logout } = useSession();

  const userName = getUserName(session);
  const userEmail = getUserEmail(session);
  const loginUrl = getLoginUrl();

  return (
    <div className="min-h-screen bg-[var(--color-background)] text-[var(--color-foreground)]">
      <div className="admin-shell-grid min-h-screen">
        <div className="mx-auto grid min-h-screen max-w-[1600px] lg:grid-cols-[280px_minmax(0,1fr)]">
          <aside className="border-b border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_82%,transparent)] px-4 py-5 backdrop-blur lg:border-r lg:border-b-0 lg:px-5 lg:py-6">
            <div className="admin-card space-y-5 p-5">
              <div className="space-y-2">
                <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
                  Shortlink Shop
                </p>
                <div>
                  <h1 className="text-xl font-semibold tracking-tight">Delivery Admin</h1>
                  <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                    Shared admin foundation inspired by the storefront UI layer.
                  </p>
                </div>
              </div>

              <div className="rounded-2xl border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-background)_65%,transparent)] px-4 py-3">
                <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-muted-foreground)]">
                  Workspace
                </p>
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span className="text-sm font-medium text-[var(--color-foreground)]">Operations</span>
                  <span className="admin-pill">Live</span>
                </div>
              </div>

              <nav className="space-y-6">
                <div className="space-y-2">
                  <p className="px-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-muted-foreground)]">
                    Navigation
                  </p>
                  <div className="space-y-2">
                    {primaryNav.map((item) => (
                      <NavLink
                        key={item.href}
                        item={item}
                        active={item.href === '/' ? pathname === '/' : pathname?.startsWith(item.href) ?? false}
                      />
                    ))}
                  </div>
                </div>

                <div className="space-y-2">
                  <p className="px-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-muted-foreground)]">
                    Tooling
                  </p>
                  <div className="space-y-2">
                    {secondaryNav.map((item) => (
                      <NavLink key={item.href} item={item} active={false} />
                    ))}
                  </div>
                </div>
              </nav>
            </div>
          </aside>

          <div className="flex min-w-0 flex-col">
            <header className="sticky top-0 z-20 border-b border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-background)_78%,transparent)] px-4 py-4 backdrop-blur sm:px-6 lg:px-8">
              <AppHeader
                className="mx-auto max-w-none px-0 py-0"
                brand={{
                  name: 'Delivery Admin',
                  href: '/',
                  render: ({ href, className }) => (
                    <HeaderLink href={href} className={className}>
                      <span className="inline-flex items-center gap-3 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-2 shadow-[0_18px_44px_-36px_rgba(15,23,42,0.26)]">
                        <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-[color-mix(in_srgb,var(--color-accent)_18%,transparent)] text-sm font-bold text-[var(--color-accent)]">
                          DA
                        </span>
                        <span className="min-w-0">
                          <span className="block text-sm font-semibold tracking-tight text-[var(--color-foreground)]">
                            Delivery Admin
                          </span>
                          <span className="block text-[11px] font-semibold uppercase tracking-[0.18em] text-[var(--color-muted-foreground)]">
                            Operations workspace
                          </span>
                        </span>
                      </span>
                    </HeaderLink>
                  ),
                }}
                workspaceLabel="Operations"
                statusBadge={{ label: 'Live', tone: 'accent' }}
                navigation={primaryNav.map((item) => ({ name: item.label, href: item.href }))}
                currentPath={pathname}
                LinkComponent={HeaderLink}
                showMenuButton={false}
                showThemeToggle={true}
                showNotifications={false}
                showProfile={hasSession && !isLoading}
                profile={
                  hasSession
                    ? {
                        name: userName,
                        email: userEmail,
                        menuItems: [
                          {
                            name: 'Django Admin',
                            href: 'https://admin.shop.shortlink.best/admin',
                          },
                          {
                            name: 'Sign out',
                            onClick: async () => {
                              await logout();
                            },
                            confirmDialog: {
                              title: 'Sign out?',
                              description: 'You will need to sign in again to access the admin panel.',
                              confirmText: 'Sign out',
                              variant: 'danger',
                            },
                          },
                        ],
                      }
                    : undefined
                }
                showLogin={!hasSession && !isLoading}
                loginButton={{ href: loginUrl, label: 'Sign in' }}
                sticky={false}
                fullWidth={true}
              />
            </header>

            <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">{children}</main>
          </div>
        </div>
      </div>
    </div>
  );
}
