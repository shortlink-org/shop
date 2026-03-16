'use client';

import { Button, FeedbackPanel, StatCard } from '@shortlink-org/ui-kit';
import Link from 'next/link';

import { DetailCard, DetailRow } from './detail-card';
import {
  formatAddress,
  formatDateTime,
  formatMoney,
  formatStatus,
  getStatusBadgeClass,
  buildMiniMap
} from './utils';
import type { DeliveryTracking, OrderState } from '@/types/order';

type ConnectionState = 'idle' | 'connecting' | 'live' | 'reconnecting';
type TrackingEvent = {
  id: string;
  title: string;
  description: string;
  timestampLabel: string;
  tone: 'neutral' | 'accent' | 'success' | 'warning' | 'danger';
};

type OrderDetailContentProps = {
  order: OrderState;
  orderId: string;
  liveTracking: DeliveryTracking | null;
  displayConnectionState: ConnectionState;
  displayedEvents: TrackingEvent[];
  onRefetch: () => void;
};

export function OrderDetailContent({
  order,
  orderId,
  liveTracking,
  displayConnectionState,
  displayedEvents,
  onRefetch
}: OrderDetailContentProps) {
  const totalItems = order?.items?.reduce((sum, item) => sum + (item.quantity ?? 0), 0) ?? 0;
  const orderTotal =
    order?.items?.reduce((sum, item) => sum + (item.quantity ?? 0) * (item.price ?? 0), 0) ?? 0;
  const deliveryAddress = order?.deliveryInfo?.deliveryAddress ?? null;
  const miniMap = buildMiniMap(liveTracking?.courier?.currentLocation, deliveryAddress);

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
            <Button variant="secondary" onClick={() => void onRefetch()}>
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
                ? 'Auto refresh'
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
              message="Tracking events appear as automatic refreshes detect courier status changes."
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
                  {order.items.map((item) => (
                    <tr
                      key={item.id ?? `item-${order.id}-${item.quantity}-${item.price}`}
                      className="border-b border-[var(--color-border)]/70 last:border-b-0"
                    >
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
