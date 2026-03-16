import { OrderDetailView } from '@/components/orders/order-detail-view';
import type { Metadata } from 'next';

interface OrderDetailPageProps {
  params: Promise<{
    id: string;
  }>;
}

export async function generateMetadata({ params }: OrderDetailPageProps): Promise<Metadata> {
  const { id } = await params;
  return {
    title: `Order ${id}`,
    description: 'Order detail and delivery tracking.'
  };
}

export default async function OrderDetailPage({ params }: OrderDetailPageProps) {
  const { id } = await params;

  return <OrderDetailView orderId={id} />;
}
