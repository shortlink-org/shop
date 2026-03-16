'use client';

import {
  CheckCircleIcon,
  ClockIcon,
  ExclamationTriangleIcon,
  MapPinIcon,
  TruckIcon,
  UserIcon
} from '@heroicons/react/24/outline';
import Link from 'next/link';
import { useEffect, useEffectEvent, useState } from 'react';

import { getDeliveryTracking } from 'lib/shopify';
import type {
  DeliveryAddress,
  DeliveryTrackingLocation,
  DeliveryTrackingSummary,
  OrderSummary
} from 'lib/shopify/types';

type StepState = 'complete' | 'current' | 'upcoming' | 'issue';
type ConnectionState = 'idle' | 'connecting' | 'live' | 'reconnecting';

type ProgressStep = {
  id: string;
  title: string;
  description: string;
  state: StepState;
};

type OrderTrackingViewProps = {
  orderId: string;
  order: OrderSummary | null;
  initialTracking: DeliveryTrackingSummary | null;
  authorization?: string;
};

const TRACKING_RANK: Record<string, number> = {
  ACCEPTED: 2,
  IN_POOL: 2,
  ASSIGNED: 3,
  IN_TRANSIT: 4,
  DELIVERED: 5
};

const ISSUE_STATUSES = new Set(['NOT_DELIVERED', 'REQUIRES_HANDLING']);
const TERMINAL_STATUSES = new Set(['DELIVERED', 'NOT_DELIVERED', 'REQUIRES_HANDLING']);

function formatStatus(status?: string | null): string {
  if (!status) return 'Processing';
  return status
    .toLowerCase()
    .split('_')
    .map((chunk) => chunk.charAt(0).toUpperCase() + chunk.slice(1))
    .join(' ');
}

function formatDateTime(value?: string | null): string | null {
  if (!value) return null;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;

  return new Intl.DateTimeFormat('en', {
    dateStyle: 'medium',
    timeStyle: 'short'
  }).format(date);
}

function formatDistanceKm(value?: number | null): string | null {
  if (typeof value !== 'number' || Number.isNaN(value)) return null;
  return `${value.toFixed(value < 10 ? 1 : 0)} km`;
}

function buildProgress(
  order: OrderSummary | null,
  tracking: DeliveryTrackingSummary | null
): ProgressStep[] {
  const trackingStatus = tracking?.status ?? undefined;

  if (trackingStatus && ISSUE_STATUSES.has(trackingStatus)) {
    return [
      {
        id: 'placed',
        title: 'Order placed',
        description: 'Your order was created successfully.',
        state: 'complete'
      },
      {
        id: 'issue',
        title: 'Delivery issue',
        description: 'The delivery requires attention. Please refresh for the latest update.',
        state: 'issue'
      }
    ];
  }

  const rank = trackingStatus ? (TRACKING_RANK[trackingStatus] ?? 1) : 1;
  const orderCompleted = order?.status === 'COMPLETED' || trackingStatus === 'DELIVERED';

  return [
    {
      id: 'placed',
      title: 'Order placed',
      description: 'We received your order and created a delivery request.',
      state: 'complete'
    },
    {
      id: 'preparing',
      title: 'Preparing order',
      description: 'Your package is being prepared for handoff.',
      state: rank > 1 ? 'complete' : 'current'
    },
    {
      id: 'assigned',
      title: 'Courier assignment',
      description: tracking?.courier
        ? `${tracking.courier.name || 'A courier'} has been assigned.`
        : 'We are finding the best courier for your delivery.',
      state: rank > 2 ? 'complete' : rank === 2 ? 'current' : 'upcoming'
    },
    {
      id: 'transit',
      title: 'On the way',
      description:
        trackingStatus === 'IN_TRANSIT'
          ? 'The courier is currently heading to your address.'
          : 'The courier will start the trip once pickup is confirmed.',
      state: orderCompleted
        ? 'complete'
        : rank === 4
          ? 'current'
          : rank > 4
            ? 'complete'
            : 'upcoming'
    },
    {
      id: 'delivered',
      title: 'Delivered',
      description: 'The order has arrived at the destination.',
      state: orderCompleted ? 'complete' : 'upcoming'
    }
  ];
}

function getBadgeClasses(status?: string | null): string {
  if (!status) return 'border-neutral-200 bg-neutral-100 text-neutral-700';
  if (status === 'DELIVERED' || status === 'COMPLETED') {
    return 'border-green-200 bg-green-100 text-green-700';
  }
  if (ISSUE_STATUSES.has(status)) {
    return 'border-amber-200 bg-amber-100 text-amber-700';
  }
  if (status === 'IN_TRANSIT' || status === 'ASSIGNED') {
    return 'border-blue-200 bg-blue-100 text-blue-700';
  }

  return 'border-neutral-200 bg-neutral-100 text-neutral-700';
}

function buildMiniMap(
  courierLocation?: DeliveryTrackingLocation | null,
  destination?: DeliveryAddress | null
) {
  const courierLat = courierLocation?.latitude;
  const courierLon = courierLocation?.longitude;
  const destinationLat = destination?.latitude;
  const destinationLon = destination?.longitude;

  if (
    typeof courierLat !== 'number' ||
    typeof courierLon !== 'number' ||
    typeof destinationLat !== 'number' ||
    typeof destinationLon !== 'number'
  ) {
    return null;
  }

  const latPadding = Math.max(Math.abs(courierLat - destinationLat) * 0.25, 0.01);
  const lonPadding = Math.max(Math.abs(courierLon - destinationLon) * 0.25, 0.01);
  const minLat = Math.min(courierLat, destinationLat) - latPadding;
  const maxLat = Math.max(courierLat, destinationLat) + latPadding;
  const minLon = Math.min(courierLon, destinationLon) - lonPadding;
  const maxLon = Math.max(courierLon, destinationLon) + lonPadding;

  const toPercentX = (lon: number) => ((lon - minLon) / (maxLon - minLon || 1)) * 100;
  const toPercentY = (lat: number) => (1 - (lat - minLat) / (maxLat - minLat || 1)) * 100;

  return {
    courier: {
      left: `${toPercentX(courierLon)}%`,
      top: `${toPercentY(courierLat)}%`
    },
    destination: {
      left: `${toPercentX(destinationLon)}%`,
      top: `${toPercentY(destinationLat)}%`
    }
  };
}

function isTerminal(status?: string | null): boolean {
  return Boolean(status && TERMINAL_STATUSES.has(status));
}

export function OrderTrackingView({
  orderId,
  order,
  initialTracking,
  authorization
}: OrderTrackingViewProps) {
  const [tracking, setTracking] = useState<DeliveryTrackingSummary | null>(initialTracking);
  const [connectionState, setConnectionState] = useState<ConnectionState>(
    authorization && !isTerminal(initialTracking?.status) ? 'connecting' : 'idle'
  );

  const applyTrackingUpdate = useEffectEvent((nextTracking: DeliveryTrackingSummary | null) => {
    setTracking(nextTracking);
  });

  useEffect(() => {
    if (!authorization || isTerminal(tracking?.status)) {
      setConnectionState('idle');
      return;
    }

    let pollTimer: ReturnType<typeof setTimeout> | undefined;
    let disposed = false;
    const poll = async () => {
      setConnectionState((current) => (current === 'live' ? 'reconnecting' : 'connecting'));
      try {
        const nextTracking = await getDeliveryTracking(orderId, { authorization });
        if (disposed) return;

        applyTrackingUpdate(nextTracking);
        setConnectionState('live');
      } catch {
        if (disposed) return;
        setConnectionState('reconnecting');
      }
      pollTimer = setTimeout(poll, 10000);
    };

    void poll();

    return () => {
      disposed = true;
      if (pollTimer) clearTimeout(pollTimer);
    };
  }, [applyTrackingUpdate, authorization, orderId, tracking?.status]);

  const deliveryAddress = order?.deliveryInfo?.deliveryAddress ?? null;
  const deliveryPeriod = order?.deliveryInfo?.deliveryPeriod ?? null;
  const steps = buildProgress(order, tracking);
  const displayStatus = tracking?.status ?? order?.status ?? null;
  const eta = tracking?.estimatedMinutesRemaining;
  const distance = formatDistanceKm(tracking?.distanceKmRemaining);
  const deliveryWindow =
    formatDateTime(deliveryPeriod?.startTime) && formatDateTime(deliveryPeriod?.endTime)
      ? `${formatDateTime(deliveryPeriod?.startTime)} - ${formatDateTime(deliveryPeriod?.endTime)}`
      : null;
  const map = buildMiniMap(tracking?.courier?.currentLocation, deliveryAddress);
  const hasIssue = Boolean(tracking?.status && ISSUE_STATUSES.has(tracking.status));

  return (
    <div className="mx-auto max-w-5xl px-4 py-12">
      <div className="rounded-[2rem] border border-[var(--color-border)] bg-[var(--color-surface)] p-6 shadow-sm sm:p-8">
        <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-2xl">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <CheckCircleIcon className="h-9 w-9 text-green-600" />
            </div>

            <h1 className="mt-6 text-3xl font-semibold tracking-tight text-[var(--color-foreground)] sm:text-4xl">
              Order confirmed
            </h1>

            <p className="mt-3 max-w-xl text-base text-[var(--color-muted-foreground)]">
              You can track preparation, courier assignment, and the latest delivery progress from
              this page.
            </p>

            <div className="mt-6 flex flex-wrap items-center gap-3">
              <span
                className={`inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium ${getBadgeClasses(displayStatus)}`}
              >
                {formatStatus(displayStatus)}
              </span>
              <span
                className={`inline-flex items-center rounded-full border px-3 py-1 text-xs font-medium ${
                  connectionState === 'live'
                    ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                    : connectionState === 'reconnecting' || connectionState === 'connecting'
                      ? 'border-blue-200 bg-blue-50 text-blue-700'
                      : 'border-neutral-200 bg-neutral-100 text-neutral-600'
                }`}
              >
                {connectionState === 'live'
                  ? 'Auto refresh'
                  : connectionState === 'reconnecting'
                    ? 'Reconnecting'
                    : connectionState === 'connecting'
                      ? 'Connecting'
                      : 'Static snapshot'}
              </span>
              <span className="text-sm text-[var(--color-muted-foreground)]">
                Order ID:{' '}
                <span className="font-mono text-[var(--color-foreground)]">{orderId}</span>
              </span>
              {tracking?.packageId ? (
                <span className="text-sm text-[var(--color-muted-foreground)]">
                  Package:{' '}
                  <span className="font-mono text-[var(--color-foreground)]">
                    {tracking.packageId}
                  </span>
                </span>
              ) : null}
            </div>
          </div>

          <div className="grid gap-3 sm:grid-cols-2 lg:w-[22rem]">
            <div className="rounded-2xl border border-[var(--color-border)] bg-[var(--color-background)] p-4">
              <div className="flex items-center gap-2 text-sm text-[var(--color-muted-foreground)]">
                <ClockIcon className="h-4 w-4" />
                ETA
              </div>
              <div className="mt-2 text-2xl font-semibold text-[var(--color-foreground)]">
                {typeof eta === 'number' ? `${eta} min` : 'Pending'}
              </div>
              <div className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                {formatDateTime(tracking?.estimatedArrivalAt) ??
                  'Shown after courier location is available'}
              </div>
            </div>

            <div className="rounded-2xl border border-[var(--color-border)] bg-[var(--color-background)] p-4">
              <div className="flex items-center gap-2 text-sm text-[var(--color-muted-foreground)]">
                <TruckIcon className="h-4 w-4" />
                Distance left
              </div>
              <div className="mt-2 text-2xl font-semibold text-[var(--color-foreground)]">
                {distance ?? 'Pending'}
              </div>
              <div className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                Based on the latest courier coordinates
              </div>
            </div>
          </div>
        </div>

        {hasIssue ? (
          <div className="mt-8 flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-amber-900">
            <ExclamationTriangleIcon className="mt-0.5 h-5 w-5 flex-shrink-0" />
            <div>
              <div className="font-medium">Delivery needs attention</div>
              <div className="text-sm text-amber-800">
                The courier could not complete the delivery. Refresh the page for the latest
                recovery status.
              </div>
            </div>
          </div>
        ) : null}

        <div className="mt-8 grid gap-8 lg:grid-cols-[1.2fr_0.8fr]">
          <div className="space-y-6">
            <section className="rounded-2xl border border-[var(--color-border)] bg-[var(--color-background)] p-5">
              <h2 className="text-lg font-semibold text-[var(--color-foreground)]">Progress</h2>
              <div className="mt-5 space-y-4">
                {steps.map((step, index) => {
                  const circleClasses =
                    step.state === 'complete'
                      ? 'bg-green-600 text-white'
                      : step.state === 'current'
                        ? 'bg-blue-600 text-white'
                        : step.state === 'issue'
                          ? 'bg-amber-500 text-white'
                          : 'bg-neutral-200 text-neutral-600';

                  return (
                    <div key={step.id} className="flex items-start gap-4">
                      <div className="flex flex-col items-center">
                        <div
                          className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold ${circleClasses}`}
                        >
                          {step.state === 'complete' ? '✓' : index + 1}
                        </div>
                        {index < steps.length - 1 ? (
                          <div className="mt-2 h-10 w-px bg-[var(--color-border)]" />
                        ) : null}
                      </div>
                      <div className="pt-1">
                        <div className="font-medium text-[var(--color-foreground)]">
                          {step.title}
                        </div>
                        <div className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                          {step.description}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>

            <section className="rounded-2xl border border-[var(--color-border)] bg-[var(--color-background)] p-5">
              <h2 className="text-lg font-semibold text-[var(--color-foreground)]">Delivery map</h2>
              {map ? (
                <div className="mt-4">
                  <div className="relative h-64 overflow-hidden rounded-2xl border border-[var(--color-border)] bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.16),transparent_34%),radial-gradient(circle_at_bottom_right,rgba(16,185,129,0.18),transparent_30%),linear-gradient(180deg,#f8fafc_0%,#eef2ff_100%)]">
                    <div className="absolute inset-0 [background-image:linear-gradient(rgba(148,163,184,0.22)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.22)_1px,transparent_1px)] [background-size:32px_32px] opacity-40" />
                    <div
                      className="absolute h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white bg-blue-600 shadow-lg"
                      style={map.courier}
                    />
                    <div
                      className="absolute h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white bg-emerald-500 shadow-lg"
                      style={map.destination}
                    />
                    <div className="absolute top-4 left-4 rounded-full bg-white/90 px-3 py-1 text-xs font-medium text-neutral-700 shadow-sm">
                      Blue: courier
                    </div>
                    <div className="absolute top-12 left-4 rounded-full bg-white/90 px-3 py-1 text-xs font-medium text-neutral-700 shadow-sm">
                      Green: destination
                    </div>
                  </div>
                  <div className="mt-3 grid gap-3 text-sm text-[var(--color-muted-foreground)] sm:grid-cols-2">
                    <div className="rounded-2xl border border-[var(--color-border)] p-3">
                      Courier coordinates:{' '}
                      <span className="font-mono text-[var(--color-foreground)]">
                        {tracking?.courier?.currentLocation?.latitude?.toFixed(5)},{' '}
                        {tracking?.courier?.currentLocation?.longitude?.toFixed(5)}
                      </span>
                    </div>
                    <div className="rounded-2xl border border-[var(--color-border)] p-3">
                      Delivery point:{' '}
                      <span className="font-mono text-[var(--color-foreground)]">
                        {deliveryAddress?.latitude?.toFixed(5)},{' '}
                        {deliveryAddress?.longitude?.toFixed(5)}
                      </span>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="mt-4 rounded-2xl border border-dashed border-[var(--color-border)] p-5 text-sm text-[var(--color-muted-foreground)]">
                  We&apos;ll show the courier position here as soon as delivery coordinates become
                  available.
                </div>
              )}
            </section>
          </div>

          <div className="space-y-6">
            <section className="rounded-2xl border border-[var(--color-border)] bg-[var(--color-background)] p-5">
              <h2 className="text-lg font-semibold text-[var(--color-foreground)]">Courier</h2>
              {tracking?.courier ? (
                <div className="mt-4 space-y-4">
                  <div className="flex items-start gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-full bg-[var(--color-muted)]">
                      <UserIcon className="h-5 w-5 text-[var(--color-foreground)]" />
                    </div>
                    <div>
                      <div className="font-medium text-[var(--color-foreground)]">
                        {tracking.courier.name || 'Assigned courier'}
                      </div>
                      <div className="text-sm text-[var(--color-muted-foreground)]">
                        {tracking.courier.transportType
                          ? formatStatus(tracking.courier.transportType)
                          : 'Transport pending'}
                      </div>
                    </div>
                  </div>

                  <div className="grid gap-3 text-sm text-[var(--color-muted-foreground)]">
                    <div className="rounded-2xl border border-[var(--color-border)] p-3">
                      Status:{' '}
                      <span className="font-medium text-[var(--color-foreground)]">
                        {formatStatus(tracking.courier.status)}
                      </span>
                    </div>
                    <div className="rounded-2xl border border-[var(--color-border)] p-3">
                      Phone:{' '}
                      <span className="font-medium text-[var(--color-foreground)]">
                        {tracking.courier.phone || 'Not shared'}
                      </span>
                    </div>
                    <div className="rounded-2xl border border-[var(--color-border)] p-3">
                      Last active:{' '}
                      <span className="font-medium text-[var(--color-foreground)]">
                        {formatDateTime(tracking.courier.lastActiveAt) ?? 'Recently updated'}
                      </span>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="mt-4 rounded-2xl border border-dashed border-[var(--color-border)] p-4 text-sm text-[var(--color-muted-foreground)]">
                  A courier has not been assigned yet. We&apos;ll show the profile here once one is
                  on the route.
                </div>
              )}
            </section>

            <section className="rounded-2xl border border-[var(--color-border)] bg-[var(--color-background)] p-5">
              <h2 className="text-lg font-semibold text-[var(--color-foreground)]">Delivery</h2>
              <div className="mt-4 space-y-3 text-sm text-[var(--color-muted-foreground)]">
                <div className="rounded-2xl border border-[var(--color-border)] p-3">
                  <div className="flex items-center gap-2 font-medium text-[var(--color-foreground)]">
                    <MapPinIcon className="h-4 w-4" />
                    Destination
                  </div>
                  <div className="mt-2">
                    {deliveryAddress
                      ? [
                          deliveryAddress.street,
                          deliveryAddress.city,
                          deliveryAddress.country
                        ]
                          .filter(Boolean)
                          .join(', ')
                      : 'Delivery address pending'}
                  </div>
                </div>
                <div className="rounded-2xl border border-[var(--color-border)] p-3">
                  <div className="font-medium text-[var(--color-foreground)]">Delivery window</div>
                  <div className="mt-2">{deliveryWindow ?? 'Time slot pending'}</div>
                </div>
                <div className="rounded-2xl border border-[var(--color-border)] p-3">
                  <div className="font-medium text-[var(--color-foreground)]">Latest update</div>
                  <div className="mt-2">
                    {formatDateTime(
                      tracking?.deliveredAt ?? tracking?.estimatedArrivalAt ?? tracking?.assignedAt
                    ) ?? 'Waiting for delivery events'}
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>

        <div className="mt-8 flex flex-col gap-3 sm:flex-row">
          <Link
            href={`/order/${orderId}`}
            className="inline-flex items-center justify-center rounded-full border border-[var(--color-border)] px-5 py-3 text-sm font-medium text-[var(--color-foreground)] hover:bg-[var(--color-muted)]"
          >
            Refresh status
          </Link>
          <Link
            href="/"
            className="inline-flex items-center justify-center rounded-full bg-blue-600 px-5 py-3 text-sm font-medium text-white hover:bg-blue-700"
          >
            Continue shopping
          </Link>
        </div>
      </div>
    </div>
  );
}
