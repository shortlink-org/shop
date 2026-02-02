import { CheckCircleIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';

interface OrderConfirmationPageProps {
  params: Promise<{
    id: string;
  }>;
}

export default async function OrderConfirmationPage({ params }: OrderConfirmationPageProps) {
  const { id: orderId } = await params;

  return (
    <div className="mx-auto max-w-2xl px-4 py-16">
      <div className="text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
          <CheckCircleIcon className="h-10 w-10 text-green-600 dark:text-green-400" />
        </div>

        <h1 className="mt-6 text-3xl font-bold text-black dark:text-white">
          Order Confirmed!
        </h1>

        <p className="mt-4 text-lg text-neutral-600 dark:text-neutral-400">
          Thank you for your order. We&apos;ve received your order and will begin processing it
          shortly.
        </p>

        <div className="mt-8 rounded-lg border border-neutral-200 bg-white p-6 dark:border-neutral-700 dark:bg-neutral-800">
          <div className="text-sm text-neutral-600 dark:text-neutral-400">Order ID</div>
          <div className="mt-1 font-mono text-lg font-semibold text-black dark:text-white">
            {orderId}
          </div>
        </div>

        <div className="mt-8 space-y-4">
          <h2 className="text-lg font-semibold text-black dark:text-white">What&apos;s Next?</h2>

          <div className="space-y-3 text-left">
            <div className="flex items-start gap-3 rounded-lg border border-neutral-200 bg-neutral-50 p-4 dark:border-neutral-700 dark:bg-neutral-800/50">
              <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-blue-600 text-xs font-bold text-white">
                1
              </div>
              <div>
                <div className="font-medium text-black dark:text-white">Order Processing</div>
                <div className="text-sm text-neutral-600 dark:text-neutral-400">
                  We&apos;re preparing your order for shipment.
                </div>
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-lg border border-neutral-200 bg-neutral-50 p-4 dark:border-neutral-700 dark:bg-neutral-800/50">
              <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-neutral-300 text-xs font-bold text-neutral-600 dark:bg-neutral-600 dark:text-neutral-300">
                2
              </div>
              <div>
                <div className="font-medium text-black dark:text-white">Courier Assignment</div>
                <div className="text-sm text-neutral-600 dark:text-neutral-400">
                  A courier will be assigned to deliver your order.
                </div>
              </div>
            </div>

            <div className="flex items-start gap-3 rounded-lg border border-neutral-200 bg-neutral-50 p-4 dark:border-neutral-700 dark:bg-neutral-800/50">
              <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-neutral-300 text-xs font-bold text-neutral-600 dark:bg-neutral-600 dark:text-neutral-300">
                3
              </div>
              <div>
                <div className="font-medium text-black dark:text-white">Delivery</div>
                <div className="text-sm text-neutral-600 dark:text-neutral-400">
                  Your order will be delivered during your selected time slot.
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-10 flex flex-col gap-4 sm:flex-row sm:justify-center">
          <Link
            href="/"
            className="inline-flex items-center justify-center rounded-full bg-blue-600 px-6 py-3 text-sm font-medium text-white hover:bg-blue-700"
          >
            Continue Shopping
          </Link>
        </div>

        <p className="mt-8 text-sm text-neutral-500 dark:text-neutral-500">
          You will receive an email confirmation with tracking details once your order is shipped.
        </p>
      </div>
    </div>
  );
}

export async function generateMetadata({ params }: OrderConfirmationPageProps) {
  const { id: orderId } = await params;

  return {
    title: `Order ${orderId} Confirmed`,
    description: 'Your order has been successfully placed.'
  };
}
