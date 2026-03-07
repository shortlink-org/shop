'use client';

import { BasketItem, Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import { useRouter } from 'next/navigation';
import { useState, useEffect, useRef } from 'react';
import CheckoutForm, { CheckoutFormData } from 'components/cart/checkout-form';
import { useCart } from 'components/cart/cart-context';
import { useCartBasket } from 'components/cart/use-cart-basket';
import { formatCartMoney } from 'components/cart/ui-kit';
import { RATE_LIMIT_MESSAGE } from 'lib/constants';
import { checkout } from 'lib/shopify';
import type { PackageInfo } from 'lib/shopify/types';
import { ArrowLeftIcon, ShoppingCartIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';

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
  const { items, handleRemoveItem, handleQuantityChange } = useCartBasket();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);
  const formErrorRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (error) {
      formErrorRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, [error]);

  const handleSubmit = async (formData: CheckoutFormData) => {
    setIsLoading(true);
    setError(null);

    try {
      const result = await checkout({
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
        <FeedbackPanel
          variant="error"
          eyebrow="Checkout"
          title="We couldn't load your cart"
          message="We'll show it when it's available again. You can keep browsing while the cart service recovers."
          action={
            <Button
              as={Link}
              asProps={{ href: '/' }}
              variant="secondary"
              icon={<ArrowLeftIcon />}
            >
              Continue shopping
            </Button>
          }
        />
      </div>
    );
  }

  // Empty cart state
  if (!cart || cart.lines.length === 0) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <FeedbackPanel
          variant="empty"
          eyebrow="Checkout"
          title="Your cart is empty"
          message="Add a few products before starting checkout."
          icon={<ShoppingCartIcon className="h-6 w-6 text-slate-500" />}
          action={
            <Button as={Link} asProps={{ href: '/' }} icon={<ArrowLeftIcon />}>
              Continue shopping
            </Button>
          }
        />
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

            <ul className="space-y-3">
              {items.map((item) => (
                <BasketItem
                  key={item.id}
                  item={item}
                  onRemove={handleRemoveItem}
                  onQuantityChange={handleQuantityChange}
                  confirmRemove={false}
                />
              ))}
            </ul>

            <div className="mt-4 space-y-2 border-t border-neutral-200 pt-4 dark:border-neutral-700">
              <div className="flex justify-between text-sm text-neutral-600 dark:text-neutral-400">
                <span>Subtotal</span>
                <span>
                  {formatCartMoney(
                    cart.cost.subtotalAmount.amount,
                    cart.cost.subtotalAmount.currencyCode
                  )}
                </span>
              </div>
              <div className="flex justify-between text-sm text-neutral-600 dark:text-neutral-400">
                <span>Taxes</span>
                <span>
                  {formatCartMoney(
                    cart.cost.totalTaxAmount?.amount ?? 0,
                    cart.cost.totalTaxAmount?.currencyCode ?? 'USD'
                  )}
                </span>
              </div>
              <div className="flex justify-between text-sm text-neutral-600 dark:text-neutral-400">
                <span>Shipping</span>
                <span>Calculated at delivery</span>
              </div>
              <div className="flex justify-between border-t border-neutral-200 pt-2 text-base font-semibold text-black dark:border-neutral-700 dark:text-white">
                <span>Total</span>
                <span>{formatCartMoney(cart.cost.totalAmount.amount, cart.cost.totalAmount.currencyCode)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Checkout Form */}
        <div className="order-1 lg:order-2">
          <div
            ref={formErrorRef}
            className="rounded-lg border border-neutral-200 bg-white p-6 dark:border-neutral-700 dark:bg-neutral-800"
          >
            <CheckoutForm onSubmit={handleSubmit} isLoading={isLoading} submitError={error} />
          </div>
        </div>
      </div>
    </div>
  );
}
