'use client';

import { Button, FeedbackPanel, StatCard } from '@shortlink-org/ui-kit';
import { useMemo, useState } from 'react';
import { useMutation, useQuery } from '@apollo/client/react';
import Link from 'next/link';
import { toast } from 'sonner';

import { CourierStatusBadge } from '@/components/couriers/CourierStatusBadge';
import { TransportBadge } from '@/components/couriers/TransportBadge';
import { GET_COURIERS } from '@/graphql/queries/couriers';
import { ACTIVATE_COURIER, DEACTIVATE_COURIER } from '@/graphql/mutations/couriers';
import type { Courier, CourierStatus, TransportType } from '@/types/courier';

const STATUS_OPTIONS: Array<{ value: CourierStatus; label: string }> = [
  { value: 'FREE', label: 'Available' },
  { value: 'BUSY', label: 'Busy' },
  { value: 'UNAVAILABLE', label: 'Unavailable' },
  { value: 'ARCHIVED', label: 'Archived' },
];

const TRANSPORT_OPTIONS: Array<{ value: TransportType; label: string }> = [
  { value: 'WALKING', label: 'Walking' },
  { value: 'BICYCLE', label: 'Bicycle' },
  { value: 'MOTORCYCLE', label: 'Motorcycle' },
  { value: 'CAR', label: 'Car' },
];

type CouriersQueryResult = {
  couriers?: {
    couriers?: Courier[];
    totalCount?: number;
  };
};

function FilterPills<T extends string>({
  label,
  options,
  selected,
  onToggle,
}: {
  label: string;
  options: Array<{ value: T; label: string }>;
  selected: T[];
  onToggle: (value: T) => void;
}) {
  return (
    <div className="space-y-3">
      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--color-muted-foreground)]">
        {label}
      </p>
      <div className="flex flex-wrap gap-2">
        {options.map((option) => {
          const active = selected.includes(option.value);
          return (
            <button
              key={option.value}
              type="button"
              className={[
                'rounded-full border px-3 py-2 text-sm font-medium transition',
                active
                  ? 'border-[var(--color-accent)] bg-[color-mix(in_srgb,var(--color-accent)_14%,transparent)] text-[var(--color-accent)]'
                  : 'border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_72%,transparent)] text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]',
              ].join(' ')}
              onClick={() => onToggle(option.value)}
            >
              {option.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}

export default function CouriersListPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState<CourierStatus[]>([]);
  const [transportFilter, setTransportFilter] = useState<TransportType[]>([]);

  const variables = useMemo(
    () => ({
      filter: {
        ...(statusFilter.length ? { statusFilter } : {}),
        ...(transportFilter.length ? { transportTypeFilter: transportFilter } : {}),
      },
      pagination: { page, pageSize },
    }),
    [page, pageSize, statusFilter, transportFilter]
  );

  const { data, loading, refetch } = useQuery<CouriersQueryResult>(GET_COURIERS, {
    variables,
    notifyOnNetworkStatusChange: true,
  });

  const [activateCourier] = useMutation(ACTIVATE_COURIER, {
    onCompleted: () => {
      toast.success('Courier activated');
      void refetch();
    },
    onError: (error) => toast.error(error.message),
  });
  const [deactivateCourier] = useMutation(DEACTIVATE_COURIER, {
    onCompleted: () => {
      toast.success('Courier deactivated');
      void refetch();
    },
    onError: (error) => toast.error(error.message),
  });

  const handleActivate = (id: string) => {
    if (!window.confirm('Activate courier?')) return;
    void activateCourier({ variables: { id } });
  };

  const handleDeactivate = (id: string) => {
    if (!window.confirm('Deactivate courier?')) return;
    void deactivateCourier({ variables: { id } });
  };

  const couriers = data?.couriers?.couriers ?? [];
  const totalCount = data?.couriers?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const availableCount = couriers.filter((courier) => courier.status === 'FREE').length;
  const busyCount = couriers.filter((courier) => courier.status === 'BUSY').length;
  const unavailableCount = couriers.filter((courier) => courier.status === 'UNAVAILABLE').length;

  const toggleStatus = (value: CourierStatus) => {
    setPage(1);
    setStatusFilter((current) =>
      current.includes(value) ? current.filter((item) => item !== value) : [...current, value]
    );
  };

  const toggleTransport = (value: TransportType) => {
    setPage(1);
    setTransportFilter((current) =>
      current.includes(value) ? current.filter((item) => item !== value) : [...current, value]
    );
  };

  return (
    <div className="space-y-6">
      <section className="admin-card overflow-hidden p-6 sm:p-8">
        <div className="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
          <div className="space-y-3">
            <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
              Courier workspace
            </p>
            <div>
              <h1 className="text-3xl font-semibold tracking-tight">Couriers</h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--color-muted-foreground)]">
                First feature slice migrated away from `refine`: direct Apollo queries, shared shell,
                `sonner` notifications, and data-first operational layout.
              </p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <Button variant="secondary" onClick={() => void refetch()}>
              Refresh
            </Button>
            <Button as={Link} asProps={{ href: '/couriers/create' }}>
              Add courier
            </Button>
          </div>
        </div>
      </section>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Visible couriers" value={couriers.length} tone="neutral" className="admin-card p-5" />
        <StatCard
          label="Available now"
          value={<span className="text-[var(--color-success)]">{availableCount}</span>}
          tone="success"
          className="admin-card p-5"
        />
        <StatCard
          label="Busy now"
          value={<span className="text-[var(--color-accent)]">{busyCount}</span>}
          tone="accent"
          className="admin-card p-5"
        />
        <StatCard
          label="Unavailable"
          value={<span className="text-[var(--color-warning)]">{unavailableCount}</span>}
          tone="warning"
          className="admin-card p-5"
        />
      </section>

      <section className="admin-card p-6">
        <div className="grid gap-6 xl:grid-cols-2">
          <FilterPills
            label="Status filters"
            options={STATUS_OPTIONS}
            selected={statusFilter}
            onToggle={toggleStatus}
          />
          <FilterPills
            label="Transport filters"
            options={TRANSPORT_OPTIONS}
            selected={transportFilter}
            onToggle={toggleTransport}
          />
        </div>
      </section>

      <section className="admin-card overflow-hidden">
        <div className="border-b border-[var(--color-border)] px-6 py-4">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-lg font-semibold tracking-tight">Fleet overview</h2>
              <p className="text-sm text-[var(--color-muted-foreground)]">
                Total matched couriers: {totalCount}
              </p>
            </div>
            <div className="flex items-center gap-3 text-sm text-[var(--color-muted-foreground)]">
              <label htmlFor="couriers-page-size">Rows</label>
              <select
                id="couriers-page-size"
                className="rounded-full border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_78%,transparent)] px-3 py-2 text-sm text-[var(--color-foreground)]"
                value={pageSize}
                onChange={(event) => {
                  setPage(1);
                  setPageSize(Number(event.target.value));
                }}
              >
                {[10, 20, 50].map((size) => (
                  <option key={size} value={size}>
                    {size}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {loading ? (
          <div className="px-6 py-16">
            <FeedbackPanel
              variant="loading"
              eyebrow="Courier workspace"
              title="Loading couriers"
              message="Fetching the latest fleet state from GraphQL."
              className="mx-auto max-w-2xl"
            />
          </div>
        ) : couriers.length === 0 ? (
          <div className="px-6 py-16">
            <FeedbackPanel
              variant="empty"
              eyebrow="Courier workspace"
              title="No couriers found"
              message="Adjust the filters or register the first courier for this workspace."
              className="mx-auto max-w-2xl"
              action={
                <Button as={Link} asProps={{ href: '/couriers/create' }}>
                  Add courier
                </Button>
              }
            />
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-[var(--color-border)] text-xs uppercase tracking-[0.18em] text-[var(--color-muted-foreground)]">
                  <th className="px-6 py-4 font-semibold">Courier</th>
                  <th className="px-6 py-4 font-semibold">Status</th>
                  <th className="px-6 py-4 font-semibold">Transport</th>
                  <th className="px-6 py-4 font-semibold">Zone</th>
                  <th className="px-6 py-4 font-semibold">Load</th>
                  <th className="px-6 py-4 font-semibold">Deliveries</th>
                  <th className="px-6 py-4 font-semibold">Rating</th>
                  <th className="px-6 py-4 font-semibold">Actions</th>
                </tr>
              </thead>
              <tbody>
                {couriers.map((courier) => (
                  <tr key={courier.courierId} className="border-b border-[var(--color-border)]/70 last:border-b-0">
                    <td className="px-6 py-5 align-top">
                      <div className="space-y-1">
                        <Link
                          href={`/couriers/${courier.courierId}`}
                          className="text-sm font-semibold text-[var(--color-foreground)] hover:text-[var(--color-accent)]"
                        >
                          {courier.name}
                        </Link>
                        <div className="text-xs text-[var(--color-muted-foreground)]">
                          {courier.email || courier.phone}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5 align-top">
                      <CourierStatusBadge status={courier.status} />
                    </td>
                    <td className="px-6 py-5 align-top">
                      <TransportBadge type={courier.transportType} />
                    </td>
                    <td className="px-6 py-5 align-top text-sm text-[var(--color-muted-foreground)]">
                      {courier.workZone}
                    </td>
                    <td className="px-6 py-5 align-top text-sm text-[var(--color-muted-foreground)]">
                      {courier.currentLoad} / {courier.maxLoad}
                    </td>
                    <td className="px-6 py-5 align-top text-sm text-[var(--color-muted-foreground)]">
                      <span className="font-semibold text-[var(--color-success)]">{courier.successfulDeliveries}</span>
                      <span className="mx-1 text-[var(--color-border)]">/</span>
                      <span className="font-semibold text-[var(--color-danger)]">{courier.failedDeliveries}</span>
                    </td>
                    <td className="px-6 py-5 align-top text-sm text-[var(--color-muted-foreground)]">
                      {courier.rating ? courier.rating.toFixed(1) : '—'}
                    </td>
                    <td className="px-6 py-5 align-top">
                      <div className="flex flex-wrap gap-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          as={Link}
                          asProps={{ href: `/couriers/${courier.courierId}` }}
                        >
                          View
                        </Button>
                        {courier.status === 'UNAVAILABLE' && (
                          <Button size="sm" onClick={() => handleActivate(courier.courierId)}>
                            Activate
                          </Button>
                        )}
                        {(courier.status === 'FREE' || courier.status === 'BUSY') && (
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleDeactivate(courier.courierId)}
                          >
                            Deactivate
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <div className="flex flex-col gap-4 border-t border-[var(--color-border)] px-6 py-4 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-[var(--color-muted-foreground)]">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-3">
            <Button
              variant="secondary"
              disabled={page <= 1}
              onClick={() => setPage((current) => Math.max(1, current - 1))}
            >
              Previous
            </Button>
            <Button
              variant="secondary"
              disabled={page >= totalPages}
              onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
            >
              Next
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
}
