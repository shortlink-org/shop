'use client';

import { Button, FeedbackPanel, StatCard } from '@shortlink-org/ui-kit';
import { useQuery } from '@apollo/client/react';
import Link from 'next/link';
import { useEffect, useEffectEvent, useState } from 'react';

import { GET_ORDER_LOOKUP } from '@/graphql/queries/orders';
import type { DeliveryAddress, DeliveryTracking, DeliveryTrackingLocation, OrderState } from '@/types/order';

type OrderLookupQueryResult = {
  getOrder?: {
    order?: OrderState | null;
  } | null;
  deliveryTracking?: DeliveryTracking | null;
};

type ConnectionState = 'idle' | 'connecting' | 'live' | 'reconnecting';
type TrackingEvent = {
  id: string;
  title: string;
  description: string;
  timestampLabel: string;
  tone: 'neutral' | 'accent' | 'success' | 'warning' | 'danger';
};

const TERMINAL_TRACKING_STATUSES = new Set(['DELIVERED', 'NOT_DELIVERED', 'REQUIRES_HANDLING']);

function formatStatus(status?: string | null): string {
  if (!status) return 'Unknown';
  return status
    .toLowerCase()
    .split('_')
    .map((chunk) => chunk.charAt(0).toUpperCase() + chunk.slice(1))
    .join(' ');
}

function formatDateTime(value?: string | null): string {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
}

function formatMoney(value?: number | null): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return '—';
  return new Intl.NumberFormat('en', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 2,
  }).format(value);
}

function formatAddress(address?: DeliveryAddress | null): string {
  if (!address) return '—';
  return [address.street, address.city, address.postalCode, address.country].filter(Boolean).join(', ') || '—';
}

function formatDateTimeShort(value?: string | null): string {
  if (!value) return 'Just now';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Just now';
  return new Intl.DateTimeFormat('en', {
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
    day: 'numeric',
  }).format(date);
}

function getStatusBadgeClass(status?: string | null): string {
  if (!status) {
    return 'border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_72%,transparent)] text-[var(--color-muted-foreground)]';
  }
  if (status === 'DELIVERED' || status === 'COMPLETED') {
    return 'border-[color-mix(in_srgb,var(--color-success)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-success)_12%,transparent)] text-[var(--color-success)]';
  }
  if (status === 'ASSIGNED' || status === 'IN_TRANSIT') {
    return 'border-[color-mix(in_srgb,var(--color-accent)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-accent)_12%,transparent)] text-[var(--color-accent)]';
  }
  if (status === 'NOT_DELIVERED' || status === 'REQUIRES_HANDLING' || status === 'CANCELLED') {
    return 'border-[color-mix(in_srgb,var(--color-danger)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-danger)_12%,transparent)] text-[var(--color-danger)]';
  }
  return 'border-[color-mix(in_srgb,var(--color-warning)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-warning)_12%,transparent)] text-[var(--color-warning)]';
}

function isTerminalTracking(status?: string | null): boolean {
  return Boolean(status && TERMINAL_TRACKING_STATUSES.has(status));
}

function getWebSocketUrl(): string {
  const configured = process.env.NEXT_PUBLIC_BFF_WS_URL;
  if (configured) return configured;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return `${protocol}//localhost:9991/graphql`;
  }

  return `${protocol}//${window.location.host}/graphql`;
}

function getEventTone(status?: string | null): TrackingEvent['tone'] {
  if (!status) return 'neutral';
  if (status === 'DELIVERED') return 'success';
  if (status === 'ASSIGNED' || status === 'IN_TRANSIT') return 'accent';
  if (status === 'NOT_DELIVERED' || status === 'REQUIRES_HANDLING' || status === 'CANCELLED') {
    return 'danger';
  }
  return 'warning';
}

function buildTrackingEvent(nextTracking: DeliveryTracking | null): TrackingEvent | null {
  if (!nextTracking) return null;

  const status = nextTracking.status ?? 'UPDATED';
  const courierName = nextTracking.courier?.name ?? 'Courier';
  const title =
    status === 'ASSIGNED'
      ? 'Courier assigned'
      : status === 'IN_TRANSIT'
        ? 'Courier on the way'
        : status === 'DELIVERED'
          ? 'Order delivered'
          : status === 'NOT_DELIVERED'
            ? 'Delivery failed'
            : status === 'REQUIRES_HANDLING'
              ? 'Delivery needs handling'
              : `Status changed to ${formatStatus(status)}`;

  const description =
    status === 'ASSIGNED'
      ? `${courierName} accepted the delivery job.`
      : status === 'IN_TRANSIT'
        ? `${courierName} is moving towards the delivery address.`
        : status === 'DELIVERED'
          ? 'The order reached the destination.'
          : status === 'NOT_DELIVERED'
            ? 'The courier could not complete delivery.'
            : status === 'REQUIRES_HANDLING'
              ? 'Manual operator attention is required.'
              : `${courierName} reported a tracking update.`;

  const timestamp =
    nextTracking.deliveredAt ??
    nextTracking.estimatedArrivalAt ??
    nextTracking.assignedAt ??
    nextTracking.courier?.lastActiveAt ??
    null;

  return {
    id: `${status}-${timestamp ?? 'now'}-${nextTracking.packageId ?? 'package'}`,
    title,
    description,
    timestampLabel: formatDateTimeShort(timestamp),
    tone: getEventTone(status),
  };
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
      top: `${toPercentY(courierLat)}%`,
    },
    destination: {
      left: `${toPercentX(destinationLon)}%`,
      top: `${toPercentY(destinationLat)}%`,
    },
  };
}

function DetailCard({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="admin-card p-6">
      <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
      <div className="mt-5 space-y-4">{children}</div>
    </section>
  );
}

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="grid gap-2 border-b border-[var(--color-border)] pb-4 last:border-b-0 last:pb-0 sm:grid-cols-[180px_minmax(0,1fr)]">
      <p className="text-sm font-medium text-[var(--color-muted-foreground)]">{label}</p>
      <div className="text-sm text-[var(--color-foreground)]">{value}</div>
    </div>
  );
}

export function OrderDetailView({ orderId }: { orderId: string }) {
  const { data, loading, refetch } = useQuery<OrderLookupQueryResult>(GET_ORDER_LOOKUP, {
    variables: { id: orderId },
    skip: !orderId,
    notifyOnNetworkStatusChange: true,
  });

  const order = data?.getOrder?.order ?? null;
  const [tracking, setTracking] = useState<DeliveryTracking | null>(null);
  const [connectionState, setConnectionState] = useState<ConnectionState>('idle');
  const [events, setEvents] = useState<TrackingEvent[]>([]);
  const liveTracking = tracking ?? data?.deliveryTracking ?? null;
  const displayConnectionState: ConnectionState =
    !orderId || isTerminalTracking(liveTracking?.status) ? 'idle' : connectionState;
  const totalItems = order?.items?.reduce((sum, item) => sum + (item.quantity ?? 0), 0) ?? 0;
  const orderTotal =
    order?.items?.reduce((sum, item) => sum + (item.quantity ?? 0) * (item.price ?? 0), 0) ?? 0;
  const deliveryAddress = order?.deliveryInfo?.deliveryAddress ?? null;
  const miniMap = buildMiniMap(liveTracking?.courier?.currentLocation, deliveryAddress);
  const displayedEvents = events.length
    ? events
    : (() => {
        const initialEvent = buildTrackingEvent(data?.deliveryTracking ?? null);
        return initialEvent ? [initialEvent] : [];
      })();

  const applyTrackingUpdate = useEffectEvent((nextTracking: DeliveryTracking | null) => {
    setTracking(nextTracking);
    const nextEvent = buildTrackingEvent(nextTracking);
    if (!nextEvent) return;
    setEvents((current) => {
      if (current[0]?.id === nextEvent.id) return current;
      return [nextEvent, ...current].slice(0, 6);
    });
  });

  useEffect(() => {
    if (!orderId || isTerminalTracking(liveTracking?.status)) {
      return;
    }

    let socket: WebSocket | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
    let disposed = false;
    const subscriptionId = `admin-order-tracking-${orderId}`;
    const authToken = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;

    const closeSocket = () => {
      if (!socket) return;
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ id: subscriptionId, type: 'complete' }));
      }
      socket.close();
      socket = null;
    };

    const connect = () => {
      setConnectionState((current) => (current === 'live' ? 'reconnecting' : 'connecting'));
      socket = new WebSocket(getWebSocketUrl(), 'graphql-transport-ws');

      socket.onopen = () => {
        socket?.send(
          JSON.stringify({
            type: 'connection_init',
            payload: authToken ? { Authorization: `Bearer ${authToken}` } : undefined,
          })
        );
      };

      socket.onmessage = (event) => {
        let message: {
          type: string;
          payload?: {
            data?: {
              deliveryTrackingUpdated?: DeliveryTracking | null;
            };
          };
        };

        try {
          message = JSON.parse(event.data);
        } catch {
          return;
        }

        if (message.type === 'connection_ack') {
          setConnectionState('live');
          socket?.send(
            JSON.stringify({
              id: subscriptionId,
              type: 'subscribe',
              payload: {
                query: `
                  subscription DeliveryTrackingUpdated($orderId: String!) {
                    deliveryTrackingUpdated(orderId: $orderId) {
                      orderId
                      packageId
                      status
                      estimatedMinutesRemaining
                      distanceKmRemaining
                      estimatedArrivalAt
                      assignedAt
                      deliveredAt
                      courier {
                        courierId
                        name
                        phone
                        transportType
                        status
                        lastActiveAt
                        currentLocation {
                          latitude
                          longitude
                        }
                      }
                    }
                  }
                `,
                variables: { orderId },
              },
            })
          );
          return;
        }

        if (message.type === 'next') {
          applyTrackingUpdate(message.payload?.data?.deliveryTrackingUpdated ?? null);
          return;
        }

        if (message.type === 'ping') {
          socket?.send(JSON.stringify({ type: 'pong' }));
          return;
        }

        if (message.type === 'complete') {
          socket?.close();
        }
      };

      socket.onerror = () => {
        socket?.close();
      };

      socket.onclose = () => {
        if (disposed || isTerminalTracking(liveTracking?.status)) {
          return;
        }

        setConnectionState('reconnecting');
        reconnectTimer = setTimeout(connect, 3000);
      };
    };

    connect();

    return () => {
      disposed = true;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      closeSocket();
    };
  }, [liveTracking?.status, orderId]);

  if (!orderId) {
    return (
      <div className="py-8">
        <FeedbackPanel
          variant="empty"
          eyebrow="Order detail"
          title="Order ID is missing"
          message="Return to order lookup and choose a valid order ID."
          className="mx-auto max-w-3xl"
          action={
            <Button as={Link} asProps={{ href: '/orders' }} variant="secondary">
              Back to lookup
            </Button>
          }
        />
      </div>
    );
  }

  if (loading) {
    return (
      <div className="py-8">
        <FeedbackPanel
          variant="loading"
          eyebrow="Order detail"
          title={`Loading order ${orderId}`}
          message="Fetching order state and delivery tracking snapshot."
          className="mx-auto max-w-3xl"
        />
      </div>
    );
  }

  if (!order) {
    return (
      <div className="py-8">
        <FeedbackPanel
          variant="empty"
          eyebrow="Order detail"
          title="Order not found"
          message={`No order payload was returned for "${orderId}". Check the ID and try again.`}
          className="mx-auto max-w-3xl"
          action={
            <Button as={Link} asProps={{ href: '/orders' }} variant="secondary">
              Back to lookup
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <section className="admin-card overflow-hidden p-6 sm:p-8">
        <div className="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
          <div className="space-y-3">
            <Button as={Link} asProps={{ href: '/orders' }} variant="secondary" size="sm">
              Back to order lookup
            </Button>
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
                Order detail
              </p>
              <h1 className="mt-2 text-3xl font-semibold tracking-tight">
                Order <span className="font-mono">{order.id ?? orderId}</span>
              </h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--color-muted-foreground)]">
                Operational view for order status, package progress, and courier assignment.
              </p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <Button variant="secondary" onClick={() => void refetch()}>
              Refresh
            </Button>
            <span
              className={`inline-flex items-center rounded-full border px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] ${
                displayConnectionState === 'live'
                  ? 'border-[color-mix(in_srgb,var(--color-success)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-success)_12%,transparent)] text-[var(--color-success)]'
                  : displayConnectionState === 'reconnecting' || displayConnectionState === 'connecting'
                    ? 'border-[color-mix(in_srgb,var(--color-accent)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-accent)_12%,transparent)] text-[var(--color-accent)]'
                    : 'border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_72%,transparent)] text-[var(--color-muted-foreground)]'
              }`}
            >
              {displayConnectionState === 'live'
                ? 'Live updates'
                : displayConnectionState === 'reconnecting'
                  ? 'Reconnecting'
                  : displayConnectionState === 'connecting'
                    ? 'Connecting'
                    : 'Static snapshot'}
            </span>
          </div>
        </div>
      </section>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Order status"
          value={
            <span
              className={`inline-flex rounded-full border px-3 py-1 text-sm font-semibold ${getStatusBadgeClass(order.status)}`}
            >
              {formatStatus(order.status)}
            </span>
          }
          tone="neutral"
          className="admin-card p-5"
        />
        <StatCard
          label="Tracking status"
          value={
            <span
              className={`inline-flex rounded-full border px-3 py-1 text-sm font-semibold ${getStatusBadgeClass(liveTracking?.status)}`}
            >
              {formatStatus(liveTracking?.status)}
            </span>
          }
          tone="neutral"
          className="admin-card p-5"
        />
        <StatCard label="Items" value={totalItems} tone="accent" className="admin-card p-5" />
        <StatCard label="Estimated total" value={formatMoney(orderTotal)} tone="success" className="admin-card p-5" />
      </section>

      <div className="grid gap-6 xl:grid-cols-2">
        <DetailCard title="Order summary">
          <DetailRow label="Order ID" value={<span className="font-mono">{order.id ?? orderId}</span>} />
          <DetailRow label="Status" value={formatStatus(order.status)} />
          <DetailRow label="Priority" value={order.deliveryInfo?.priority ?? '—'} />
          <DetailRow
            label="Delivery window"
            value={
              order.deliveryInfo?.deliveryPeriod
                ? `${order.deliveryInfo.deliveryPeriod.startTime ?? '—'} - ${order.deliveryInfo.deliveryPeriod.endTime ?? '—'}`
                : '—'
            }
          />
        </DetailCard>

        <DetailCard title="Recipient">
          <DetailRow label="Name" value={order.deliveryInfo?.recipientContacts?.recipientName ?? '—'} />
          <DetailRow label="Phone" value={order.deliveryInfo?.recipientContacts?.recipientPhone ?? '—'} />
          <DetailRow label="Email" value={order.deliveryInfo?.recipientContacts?.recipientEmail ?? '—'} />
        </DetailCard>

        <DetailCard title="Addresses">
          <DetailRow label="Pickup address" value={formatAddress(order.deliveryInfo?.pickupAddress)} />
          <DetailRow label="Delivery address" value={formatAddress(order.deliveryInfo?.deliveryAddress)} />
        </DetailCard>

        <DetailCard title="Delivery tracking">
          <DetailRow label="Package ID" value={liveTracking?.packageId ?? '—'} />
          <DetailRow label="Assigned at" value={formatDateTime(liveTracking?.assignedAt)} />
          <DetailRow label="Estimated arrival" value={formatDateTime(liveTracking?.estimatedArrivalAt)} />
          <DetailRow
            label="Remaining distance"
            value={
              typeof liveTracking?.distanceKmRemaining === 'number'
                ? `${liveTracking.distanceKmRemaining.toFixed(liveTracking.distanceKmRemaining < 10 ? 1 : 0)} km`
                : '—'
            }
          />
          <DetailRow
            label="Remaining time"
            value={
              typeof liveTracking?.estimatedMinutesRemaining === 'number'
                ? `${liveTracking.estimatedMinutesRemaining} min`
                : '—'
            }
          />
        </DetailCard>

        <DetailCard title="Assigned courier">
          {liveTracking?.courier ? (
            <>
              <DetailRow label="Courier" value={liveTracking.courier.name ?? '—'} />
              <DetailRow
                label="Courier ID"
                value={<span className="font-mono">{liveTracking.courier.courierId ?? '—'}</span>}
              />
              <DetailRow label="Phone" value={liveTracking.courier.phone ?? '—'} />
              <DetailRow label="Transport" value={liveTracking.courier.transportType ?? '—'} />
              <DetailRow label="Courier status" value={formatStatus(liveTracking.courier.status)} />
              <DetailRow label="Last active" value={formatDateTime(liveTracking.courier.lastActiveAt)} />
            </>
          ) : (
            <FeedbackPanel
              variant="empty"
              eyebrow="Delivery tracking"
              title="No courier assigned"
              message="The order has not been assigned to a courier yet, or tracking data is unavailable."
              size="sm"
            />
          )}
        </DetailCard>

        <DetailCard title="Last events">
          {displayedEvents.length ? (
            <div className="space-y-3">
              {displayedEvents.map((event) => (
                <div
                  key={event.id}
                  className="rounded-2xl border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_72%,transparent)] p-4"
                >
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <span
                        className={`inline-flex h-2.5 w-2.5 rounded-full ${
                          event.tone === 'success'
                            ? 'bg-[var(--color-success)]'
                            : event.tone === 'accent'
                              ? 'bg-[var(--color-accent)]'
                              : event.tone === 'danger'
                                ? 'bg-[var(--color-danger)]'
                                : event.tone === 'warning'
                                  ? 'bg-[var(--color-warning)]'
                                  : 'bg-[var(--color-muted-foreground)]'
                        }`}
                      />
                      <p className="text-sm font-semibold text-[var(--color-foreground)]">{event.title}</p>
                    </div>
                    <span className="text-xs font-medium uppercase tracking-[0.14em] text-[var(--color-muted-foreground)]">
                      {event.timestampLabel}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-[var(--color-muted-foreground)]">{event.description}</p>
                </div>
              ))}
            </div>
          ) : (
            <FeedbackPanel
              variant="empty"
              eyebrow="Live events"
              title="No tracking events yet"
              message="Once courier status changes start arriving over subscription, they will appear here."
              size="sm"
            />
          )}
        </DetailCard>

        <DetailCard title="Live courier location">
          {miniMap ? (
            <div className="space-y-4">
              <div className="relative h-64 overflow-hidden rounded-2xl border border-[var(--color-border)] bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.16),transparent_34%),radial-gradient(circle_at_bottom_right,rgba(16,185,129,0.18),transparent_30%),linear-gradient(180deg,#f8fafc_0%,#eef2ff_100%)]">
                <div className="absolute inset-0 [background-image:linear-gradient(rgba(148,163,184,0.22)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.22)_1px,transparent_1px)] [background-size:32px_32px] opacity-40" />
                <div
                  className="absolute h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white bg-[var(--color-accent)] shadow-lg"
                  style={miniMap.courier}
                />
                <div
                  className="absolute h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white bg-[var(--color-success)] shadow-lg"
                  style={miniMap.destination}
                />
                <div className="absolute top-4 left-4 rounded-full bg-white/90 px-3 py-1 text-xs font-medium text-neutral-700 shadow-sm">
                  Blue: courier
                </div>
                <div className="absolute top-12 left-4 rounded-full bg-white/90 px-3 py-1 text-xs font-medium text-neutral-700 shadow-sm">
                  Green: destination
                </div>
              </div>

              <div className="grid gap-3 text-sm text-[var(--color-muted-foreground)] sm:grid-cols-2">
                <div className="rounded-2xl border border-[var(--color-border)] p-3">
                  Courier coordinates:{' '}
                  <span className="font-mono text-[var(--color-foreground)]">
                    {liveTracking?.courier?.currentLocation?.latitude?.toFixed(5)},{' '}
                    {liveTracking?.courier?.currentLocation?.longitude?.toFixed(5)}
                  </span>
                </div>
                <div className="rounded-2xl border border-[var(--color-border)] p-3">
                  Delivery point:{' '}
                  <span className="font-mono text-[var(--color-foreground)]">
                    {deliveryAddress?.latitude?.toFixed(5)}, {deliveryAddress?.longitude?.toFixed(5)}
                  </span>
                </div>
              </div>
            </div>
          ) : (
            <FeedbackPanel
              variant="empty"
              eyebrow="Live map"
              title="Map coordinates are not available yet"
              message="We need both courier live coordinates and delivery destination coordinates to render the mini-map."
              size="sm"
            />
          )}
        </DetailCard>

        <DetailCard title="Order items">
          {order.items?.length ? (
            <div className="overflow-x-auto">
              <table className="min-w-full border-collapse text-left">
                <thead>
                  <tr className="border-b border-[var(--color-border)] text-xs uppercase tracking-[0.18em] text-[var(--color-muted-foreground)]">
                    <th className="px-0 py-3 font-semibold">Item ID</th>
                    <th className="px-4 py-3 font-semibold">Qty</th>
                    <th className="px-4 py-3 font-semibold">Price</th>
                    <th className="px-4 py-3 font-semibold">Line total</th>
                  </tr>
                </thead>
                <tbody>
                  {order.items.map((item, index) => (
                    <tr key={`${item.id ?? 'item'}-${index}`} className="border-b border-[var(--color-border)]/70 last:border-b-0">
                      <td className="px-0 py-3 font-mono text-sm">{item.id ?? '—'}</td>
                      <td className="px-4 py-3 text-sm">{item.quantity ?? '—'}</td>
                      <td className="px-4 py-3 text-sm">{formatMoney(item.price)}</td>
                      <td className="px-4 py-3 text-sm">
                        {formatMoney((item.quantity ?? 0) * (item.price ?? 0))}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <FeedbackPanel
              variant="empty"
              eyebrow="Order payload"
              title="No items returned"
              message="The order query returned no line items for this order."
              size="sm"
            />
          )}
        </DetailCard>
      </div>
    </div>
  );
}
