'use client';

import { Button, FeedbackPanel, StatCard } from '@shortlink-org/ui-kit';
import { useMutation, useQuery } from '@apollo/client/react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { toast } from 'sonner';

import { CourierStatusBadge } from '@/components/couriers/CourierStatusBadge';
import { TransportBadge } from '@/components/couriers/TransportBadge';
import { GET_COURIER } from '@/graphql/queries/couriers';
import { ACTIVATE_COURIER, DEACTIVATE_COURIER, ARCHIVE_COURIER } from '@/graphql/mutations/couriers';
import type { Courier } from '@/types/courier';

const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
type CourierQueryResult = {
  courier?: Courier | null;
};

function DetailCard({
  title,
  children,
  className = '',
}: {
  title: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={`admin-card p-6 ${className}`}>
      <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
      <div className="mt-5 space-y-4">{children}</div>
    </section>
  );
}

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="grid gap-2 border-b border-[var(--color-border)] pb-4 last:border-b-0 last:pb-0 sm:grid-cols-[160px_minmax(0,1fr)]">
      <p className="text-sm font-medium text-[var(--color-muted-foreground)]">{label}</p>
      <div className="text-sm text-[var(--color-foreground)]">{value}</div>
    </div>
  );
}

export default function CourierDetailPage() {
  const params = useParams();
  const courierId = params.id as string;

  const { data, loading, refetch } = useQuery<CourierQueryResult>(GET_COURIER, {
    variables: { id: courierId, includeLocation: true },
    skip: !courierId,
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
  const [archiveCourier] = useMutation(ARCHIVE_COURIER, {
    onCompleted: () => {
      toast.success('Courier archived');
      void refetch();
    },
    onError: (error) => toast.error(error.message),
  });

  const handleActivate = () => {
    if (!window.confirm('Activate courier?')) return;
    void activateCourier({ variables: { id: courierId } });
  };

  const handleDeactivate = () => {
    if (!window.confirm('Deactivate courier?')) return;
    void deactivateCourier({ variables: { id: courierId } });
  };

  const handleArchive = () => {
    if (!window.confirm('Archive courier? This action cannot be undone.')) return;
    void archiveCourier({ variables: { id: courierId } });
  };

  if (loading) {
    return (
      <div className="py-16">
        <FeedbackPanel
          variant="loading"
          eyebrow="Courier profile"
          title="Loading courier"
          message="Fetching courier details, workload, and current location."
          className="mx-auto max-w-2xl"
        />
      </div>
    );
  }

  const courier = data?.courier ?? undefined;

  if (!courier) {
    return (
      <div className="py-16">
        <FeedbackPanel
          variant="empty"
          eyebrow="Courier profile"
          title="Courier not found"
          message="The requested courier is unavailable or does not exist anymore."
          className="mx-auto max-w-2xl"
          action={
            <Button as={Link} asProps={{ href: '/couriers' }} variant="secondary">
              Back to list
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <section className="admin-card overflow-hidden p-6 sm:p-8">
        <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-3">
              <Button as={Link} asProps={{ href: '/couriers' }} variant="secondary" size="sm">
                Back to list
              </Button>
              <CourierStatusBadge status={courier.status} />
            </div>
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
                Courier profile
              </p>
              <h1 className="mt-2 text-3xl font-semibold tracking-tight">{courier.name}</h1>
              <p className="mt-2 text-sm text-[var(--color-muted-foreground)]">
                {courier.email} · {courier.phone}
              </p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <Button variant="secondary" onClick={() => void refetch()}>
              Refresh
            </Button>
          {courier.status === 'UNAVAILABLE' && (
            <Button onClick={handleActivate}>Activate</Button>
          )}
          {(courier.status === 'FREE' || courier.status === 'BUSY') && (
            <Button variant="secondary" onClick={handleDeactivate}>
              Deactivate
            </Button>
          )}
          {courier.status !== 'ARCHIVED' && (
            <Button variant="destructive" onClick={handleArchive}>
              Archive
            </Button>
          )}
          </div>
        </div>
      </section>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Rating" value={courier.rating ? courier.rating.toFixed(1) : '—'} tone="accent" className="admin-card p-5" />
        <StatCard
          label="Current load"
          value={`${courier.currentLoad} / ${courier.maxLoad}`}
          tone="neutral"
          className="admin-card p-5"
        />
        <StatCard
          label="Successful deliveries"
          value={<span className="text-[var(--color-success)]">{courier.successfulDeliveries}</span>}
          tone="success"
          className="admin-card p-5"
        />
        <StatCard
          label="Failed deliveries"
          value={<span className="text-[var(--color-danger)]">{courier.failedDeliveries}</span>}
          tone="danger"
          className="admin-card p-5"
        />
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <DetailCard title="Basic information">
          <DetailRow label="ID" value={courier.courierId} />
          <DetailRow label="Name" value={courier.name} />
          <DetailRow label="Phone" value={<a href={`tel:${courier.phone}`}>{courier.phone}</a>} />
          <DetailRow label="Email" value={<a href={`mailto:${courier.email}`}>{courier.email}</a>} />
        </DetailCard>

        <DetailCard title="Transport and work">
          <DetailRow label="Transport" value={<TransportBadge type={courier.transportType} />} />
          <DetailRow label="Max distance" value={`${courier.maxDistanceKm} km`} />
          <DetailRow label="Load" value={`${courier.currentLoad} / ${courier.maxLoad} packages`} />
          <DetailRow label="Zone" value={courier.workZone} />
          {courier.workHours && (
            <>
              <DetailRow
                label="Working hours"
                value={`${courier.workHours.startTime} - ${courier.workHours.endTime}`}
              />
              <DetailRow
                label="Working days"
                value={courier.workHours.workDays?.map((day: number) => WEEKDAYS[day]).join(', ') || '—'}
              />
            </>
          )}
        </DetailCard>

        <DetailCard title="Operational history">
          <DetailRow
            label="Registration date"
            value={courier.createdAt ? new Date(courier.createdAt).toLocaleDateString('en-US') : '—'}
          />
          <DetailRow
            label="Last activity"
            value={courier.lastActiveAt ? new Date(courier.lastActiveAt).toLocaleString('en-US') : '—'}
          />
          <DetailRow
            label="Successful deliveries"
            value={<span className="font-semibold text-[var(--color-success)]">{courier.successfulDeliveries}</span>}
          />
          <DetailRow
            label="Failed deliveries"
            value={<span className="font-semibold text-[var(--color-danger)]">{courier.failedDeliveries}</span>}
          />
        </DetailCard>

        <DetailCard title="Current location">
          {courier.currentLocation ? (
            <>
              <DetailRow label="Latitude" value={courier.currentLocation.latitude?.toFixed(6)} />
              <DetailRow label="Longitude" value={courier.currentLocation.longitude?.toFixed(6)} />
              <div className="rounded-2xl border border-dashed border-[var(--color-border)] px-4 py-8 text-center text-sm text-[var(--color-muted-foreground)]">
                Map visualization will be added in the next admin iteration.
              </div>
            </>
          ) : (
            <FeedbackPanel
              variant="empty"
              eyebrow="Courier profile"
              title="No live location"
              message="This courier has not reported a current position yet."
              size="sm"
            />
          )}
        </DetailCard>
      </div>

      <DetailCard title="Recent deliveries">
        <FeedbackPanel
          variant="empty"
          eyebrow="Delivery integration"
          title="Delivery feed is not wired yet"
          message="Courier delivery history will be shown here once the related GraphQL API is connected."
          size="sm"
        />
      </DetailCard>
    </div>
  );
}
