'use client';

import { useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import CheckoutForm, { CheckoutFormData } from 'components/cart/checkout-form';
import Price from 'components/price';
import { useCart } from 'components/cart/cart-context';
import { RATE_LIMIT_MESSAGE } from 'lib/constants';
import { checkout } from 'lib/shopify';
import type { PackageInfo } from 'lib/shopify/types';
import { ArrowLeftIcon, ShoppingCartIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';
import Image from 'next/image';

// Default warehouse/pickup address
const PICKUP_ADDRESS = {
  street: '100 Warehouse Street',
  city: 'Berlin',
  postalCode: '10115',
  country: 'Germany',
  latitude: 52.52,
  longitude: 13.405
};

// Default package info (can be calculated from cart items later)
const DEFAULT_PACKAGE_INFO: PackageInfo = {
  weightKg: 1.0
};

export default function CheckoutPage() {
  const router = useRouter();
  const { cart, cartUnavailable } = useCart();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const handleSubmit = async (formData: CheckoutFormData) => {
    if (!cart?.id) {
      setError('No cart found. Please add items to your cart first.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const result = await checkout({
        customerId: cart.id,
        deliveryInfo: {
          pickupAddress: PICKUP_ADDRESS,
          deliveryAddress: formData.deliveryAddress,
          deliveryPeriod: formData.deliveryPeriod,
          packageInfo: DEFAULT_PACKAGE_INFO,
          priority: formData.priority,
          recipientContacts: {
            recipientName: formData.recipientContacts.recipientName || undefined,
            recipientPhone: formData.recipientContacts.recipientPhone || undefined,
            recipientEmail: formData.recipientContacts.recipientEmail || undefined
          }
        }
      });

      if (result.orderId) {
        router.push(`/order/${result.orderId}`);
      } else {
        setError('Failed to create order. Please try again.');
      }
    } catch (err) {
      console.error('Checkout error:', err);
      const status = err && typeof err === 'object' && 'status' in err ? (err as { status: number }).status : undefined;
      const message = err && typeof err === 'object' && 'message' in err ? (err as { message: string }).message : undefined;
      setError(
        status === 429 || message === RATE_LIMIT_MESSAGE
          ? RATE_LIMIT_MESSAGE
          : 'An error occurred during checkout. Please try again.'
      );
    } finally {
      setIsLoading(false);
    }
  };

  // Don't render until mounted to avoid hydration mismatch
  if (!mounted) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <div className="animate-pulse">
          <div className="mb-8 h-8 w-48 rounded bg-neutral-200 dark:bg-neutral-700" />
          <div className="grid gap-8 lg:grid-cols-2">
            <div className="h-96 rounded-lg bg-neutral-200 dark:bg-neutral-700" />
            <div className="h-96 rounded-lg bg-neutral-200 dark:bg-neutral-700" />
          </div>
        </div>
      </div>
    );
  }

  // Cart service unavailable
  if (cartUnavailable) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <div className="flex flex-col items-center justify-center py-16">
          <ShoppingCartIcon className="h-16 w-16 text-neutral-400" />
          <h2 className="mt-4 text-xl font-semibold text-black dark:text-white">
            We couldn&apos;t load your cart
          </h2>
          <p className="mt-2 text-center text-neutral-600 dark:text-neutral-400">
            We&apos;ll show it when it&apos;s available again. You can keep browsing.
          </p>
          <Link
            href="/"
            className="mt-6 inline-flex items-center gap-2 rounded-full border border-neutral-300 bg-white px-6 py-2 text-sm font-medium dark:border-neutral-700 dark:bg-neutral-800"
          >
            <ArrowLeftIcon className="h-4 w-4" />
            Continue shopping
          </Link>
        </div>
      </div>
    );
  }

  // Empty cart state
  if (!cart || cart.lines.length === 0) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <div className="flex flex-col items-center justify-center py-16">
          <ShoppingCartIcon className="h-16 w-16 text-neutral-400" />
          <h2 className="mt-4 text-xl font-semibold text-black dark:text-white">
            Your cart is empty
          </h2>
          <p className="mt-2 text-neutral-600 dark:text-neutral-400">
            Add some items to your cart before checking out.
          </p>
          <Link
            href="/"
            className="mt-6 rounded-full bg-blue-600 px-6 py-3 text-sm font-medium text-white hover:bg-blue-700"
          >
            Continue Shopping
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <Link
          href="/"
          className="mb-4 inline-flex items-center text-sm text-neutral-600 hover:text-black dark:text-neutral-400 dark:hover:text-white"
        >
          <ArrowLeftIcon className="mr-2 h-4 w-4" />
          Continue Shopping
        </Link>
        <h1 className="text-2xl font-bold text-black dark:text-white">Checkout</h1>
      </div>

      {error && (
        <div className="mb-6 rounded-lg bg-red-50 p-4 text-red-700 dark:bg-red-900/20 dark:text-red-400">
          {error}
        </div>
      )}

      <div className="grid gap-8 lg:grid-cols-2">
        {/* Order Summary */}
        <div className="order-2 lg:order-1">
          <div className="rounded-lg border border-neutral-200 bg-white p-6 dark:border-neutral-700 dark:bg-neutral-800">
            <h2 className="mb-4 text-lg font-semibold text-black dark:text-white">
              Order Summary
            </h2>

            <ul className="divide-y divide-neutral-200 dark:divide-neutral-700">
              {cart.lines.map((item, i) => (
                <li key={i} className="flex py-4">
                  <div className="relative h-16 w-16 flex-shrink-0 overflow-hidden rounded-md border border-neutral-200 dark:border-neutral-700">
                    <Image
                      src="https://picsum.photos/200"
                      alt="Product"
                      width={64}
                      height={64}
                      className="h-full w-full object-cover"
                    />
                  </div>
                  <div className="ml-4 flex flex-1 flex-col">
                    <div className="flex justify-between text-sm font-medium text-black dark:text-white">
                      <span>{item.merchandise.product.title}</span>
                      <Price
                        amount={item.cost.totalAmount.amount}
                        currencyCode={item.cost.totalAmount.currencyCode}
                      />
                    </div>
                    <p className="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
                      Qty: {item.quantity}
                    </p>
                  </div>
                </li>
              ))}
            </ul>

            <div className="mt-4 space-y-2 border-t border-neutral-200 pt-4 dark:border-neutral-700">
              <div className="flex justify-between text-sm text-neutral-600 dark:text-neutral-400">
                <span>Subtotal</span>
                <Price
                  amount={cart.cost.subtotalAmount.amount}
                  currencyCode={cart.cost.subtotalAmount.currencyCode}
                />
              </div>
              <div className="flex justify-between text-sm text-neutral-600 dark:text-neutral-400">
                <span>Taxes</span>
                <Price
                  amount={cart.cost.totalTaxAmount?.amount ?? 0}
                  currencyCode={cart.cost.totalTaxAmount?.currencyCode ?? 'USD'}
                />
              </div>
              <div className="flex justify-between text-sm text-neutral-600 dark:text-neutral-400">
                <span>Shipping</span>
                <span>Calculated at delivery</span>
              </div>
              <div className="flex justify-between border-t border-neutral-200 pt-2 text-base font-semibold text-black dark:border-neutral-700 dark:text-white">
                <span>Total</span>
                <Price
                  amount={cart.cost.totalAmount.amount}
                  currencyCode={cart.cost.totalAmount.currencyCode}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Checkout Form */}
        <div className="order-1 lg:order-2">
          <div className="rounded-lg border border-neutral-200 bg-white p-6 dark:border-neutral-700 dark:bg-neutral-800">
            <CheckoutForm onSubmit={handleSubmit} isLoading={isLoading} />
          </div>
        </div>
      </div>
    </div>
  );
}
