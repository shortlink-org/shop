import { headers } from 'next/headers';

import { getOrderTrackingPage } from 'lib/shopify';

import { OrderTrackingView } from './order-tracking-view';

interface OrderConfirmationPageProps {
  params: Promise<{
    id: string;
  }>;
}

export default async function OrderConfirmationPage({ params }: OrderConfirmationPageProps) {
  const { id: orderId } = await params;
  const authorization = (await headers()).get('authorization') ?? undefined;
  const data = await getOrderTrackingPage(orderId, { authorization });

  return (
    <OrderTrackingView
      key={orderId}
      orderId={orderId}
      order={data?.order ?? null}
      initialTracking={data?.tracking ?? null}
      authorization={authorization}
    />
  );
}

export async function generateMetadata({ params }: OrderConfirmationPageProps) {
  const { id: orderId } = await params;

  return {
    title: `Order ${orderId} tracking`,
    description: 'Track order processing, courier assignment, and live delivery progress.'
  };
}
