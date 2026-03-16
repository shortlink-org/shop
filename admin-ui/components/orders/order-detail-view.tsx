'use client';

import { Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import { useQuery } from '@apollo/client/react';
import Link from 'next/link';
import { useEffect, useReducer } from 'react';

import { GET_ORDER_LOOKUP } from '@/graphql/queries/orders';
import type { DeliveryTracking, OrderState } from '@/types/order';
import { OrderDetailContent } from './order-detail/order-detail-content';
import { formatDateTimeShort, formatStatus, isTerminalTracking } from './order-detail/utils';

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

type OrderDetailState = {
  tracking: DeliveryTracking | null;
  connectionState: ConnectionState;
  events: TrackingEvent[];
};

type OrderDetailAction =
  | { type: 'TRACKING'; payload: DeliveryTracking | null }
  | { type: 'CONNECTION'; payload: ConnectionState }
  | { type: 'EVENT'; payload: TrackingEvent };

function orderDetailReducer(state: OrderDetailState, action: OrderDetailAction): OrderDetailState {
  switch (action.type) {
    case 'TRACKING':
      return { ...state, tracking: action.payload };
    case 'CONNECTION':
      return { ...state, connectionState: action.payload };
    case 'EVENT': {
      const nextEvent = action.payload;
      const updated =
        state.events[0]?.id === nextEvent.id ? state.events : [nextEvent, ...state.events].slice(0, 6);
      return { ...state, events: updated };
    }
    default:
      return state;
  }
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
    tone: getEventTone(status)
  };
}

export function OrderDetailView({ orderId }: { orderId: string }) {
  const { data, loading, refetch, startPolling, stopPolling } = useQuery<OrderLookupQueryResult>(
    GET_ORDER_LOOKUP,
    {
      variables: { id: orderId },
      skip: !orderId,
      notifyOnNetworkStatusChange: true
    }
  );

  const order = data?.getOrder?.order ?? null;
  const [state, dispatch] = useReducer(orderDetailReducer, {
    tracking: null,
    connectionState: 'idle',
    events: []
  });
  const liveTracking = state.tracking ?? data?.deliveryTracking ?? null;
  const displayConnectionState: ConnectionState =
    !orderId || isTerminalTracking(liveTracking?.status) ? 'idle' : state.connectionState;
  const displayedEvents = state.events.length
    ? state.events
    : (() => {
        const initialEvent = buildTrackingEvent(data?.deliveryTracking ?? null);
        return initialEvent ? [initialEvent] : [];
      })();

  useEffect(() => {
    const nextTracking = data?.deliveryTracking ?? null;
    dispatch({ type: 'TRACKING', payload: nextTracking });
    const nextEvent = buildTrackingEvent(nextTracking);
    if (nextEvent) {
      dispatch({ type: 'EVENT', payload: nextEvent });
    }
  }, [data?.deliveryTracking]);

  useEffect(() => {
    if (!orderId || isTerminalTracking(liveTracking?.status)) {
      stopPolling();
      queueMicrotask(() => dispatch({ type: 'CONNECTION', payload: 'idle' }));
      return;
    }

    queueMicrotask(() => dispatch({ type: 'CONNECTION', payload: loading ? 'connecting' : 'live' }));
    startPolling(10000);

    return () => {
      stopPolling();
    };
  }, [liveTracking?.status, loading, orderId, startPolling, stopPolling]);

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
    <OrderDetailContent
      order={order}
      orderId={orderId}
      liveTracking={liveTracking}
      displayConnectionState={displayConnectionState}
      displayedEvents={displayedEvents}
      onRefetch={refetch}
    />
  );
}
