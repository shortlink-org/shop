'use client';

import { Button, StatCard } from '@shortlink-org/ui-kit';
import Link from 'next/link';

const stats = [
  { label: 'Total couriers', value: 45, tone: 'text-[var(--color-foreground)]' },
  { label: 'Available now', value: 32, tone: 'text-[var(--color-success)]' },
  { label: 'On delivery', value: 18, tone: 'text-[var(--color-accent)]' },
  { label: 'Deliveries today', value: 156, tone: 'text-[var(--color-warning)]' },
];

const highlights = [
  'The root layout no longer depends on `refine` resources or router providers.',
  'Notifications and session handling now match the storefront architecture more closely.',
  'Couriers is the first module being migrated to plain Apollo and shared page primitives.',
];

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <section className="admin-card overflow-hidden p-6 sm:p-8">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1.5fr)_minmax(280px,0.9fr)] lg:items-end">
          <div className="space-y-4">
            <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
              Admin rebuild
            </p>
            <div className="space-y-3">
              <h1 className="max-w-3xl text-3xl font-semibold tracking-tight sm:text-4xl">
                Delivery operations are moving onto the shared UI platform.
              </h1>
              <p className="max-w-2xl text-sm leading-6 text-[var(--color-muted-foreground)] sm:text-base">
                This first slice replaces the `refine` shell with a storefront-inspired layout, theme,
                notifications, and admin navigation. The next step is to finish migrating the courier
                workflows page by page.
              </p>
            </div>
          </div>

          <div className="admin-card grid gap-3 p-5">
            <p className="text-sm font-semibold">Quick actions</p>
            <Button as={Link} asProps={{ href: '/orders' }} variant="secondary">
              Open order lookup
            </Button>
            <Button as={Link} asProps={{ href: '/couriers' }}>
              Open couriers workspace
            </Button>
            <Button
              variant="secondary"
              as="a"
              asProps={{
                href: 'https://admin.shop.shortlink.best/admin',
                target: '_blank',
                rel: 'noreferrer',
              }}
            >
              Open Django admin
            </Button>
          </div>
        </div>
      </section>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {stats.map((item) => (
          <StatCard
            key={item.label}
            label={item.label}
            value={<span className={item.tone}>{item.value}</span>}
            tone="neutral"
            className="admin-card p-5"
          />
        ))}
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1.4fr)_minmax(320px,0.8fr)]">
        <article className="admin-card p-6">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-muted-foreground)]">
                Migration status
              </p>
              <h2 className="mt-2 text-xl font-semibold tracking-tight">Current foundation changes</h2>
            </div>
            <span className="admin-pill">Phase 1</span>
          </div>
          <ul className="mt-6 space-y-3 text-sm leading-6 text-[var(--color-muted-foreground)]">
            {highlights.map((item) => (
              <li key={item} className="rounded-2xl border border-[var(--color-border)] px-4 py-3">
                {item}
              </li>
            ))}
          </ul>
        </article>

        <article className="admin-card p-6">
          <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-[var(--color-muted-foreground)]">
            Next up
          </p>
          <h2 className="mt-2 text-xl font-semibold tracking-tight">Recommended sequence</h2>
          <ol className="mt-6 space-y-4 text-sm leading-6 text-[var(--color-muted-foreground)]">
            <li className="rounded-2xl border border-[var(--color-border)] px-4 py-3">
              Finish `couriers` detail and create/edit screens on the new primitives.
            </li>
            <li className="rounded-2xl border border-[var(--color-border)] px-4 py-3">
              Introduce shared admin components: filters, tables, cards, empty/error states.
            </li>
            <li className="rounded-2xl border border-[var(--color-border)] px-4 py-3">
              Expand the same shared UI platform approach to orders, customers, and operational workflows.
            </li>
          </ol>
        </article>
      </section>
    </div>
  );
}
