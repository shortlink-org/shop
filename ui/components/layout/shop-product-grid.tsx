'use client';

import { ProductGrid, ProductQuickView } from '@shortlink-org/ui-kit';
import clsx from 'clsx';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { addItem } from 'components/cart/actions';
import { useCart } from 'components/cart/cart-context';
import { DEFAULT_OPTION } from 'lib/constants';
import { Good, GoodVariant } from 'lib/shopify/types';
import { toast } from 'sonner';
import type { ProductGridProduct, ProductQuickViewProduct } from '@shortlink-org/ui-kit';

const PLACEHOLDER_IMAGE = 'https://picsum.photos/400/500';
const ADDING_BADGE = [{ label: 'Adding...', tone: 'info' as const }];

function formatPrice(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

function goodToOptimisticVariant(good: Good): GoodVariant {
  return {
    id: good.id,
    title: DEFAULT_OPTION,
    availableForSale: true,
    selectedOptions: [],
    price: { amount: good.price, currencyCode: 'USD' }
  };
}

function goodToQuickViewProduct(good: Good): ProductQuickViewProduct {
  return {
    name: good.name,
    imageSrc: PLACEHOLDER_IMAGE,
    imageAlt: good.name,
    price: formatPrice(good.price),
    rating: 4.5,
    reviewCount: 24,
    reviewsHref: `/good/${good.id}#reviews`,
    colors: [],
    sizes: []
  };
}

function goodToProduct(
  good: Good,
  onAddToCart: (good: Good) => void,
  onQuickView: (good: Good) => void,
  isAdding: boolean
): ProductGridProduct {
  return {
    id: good.id,
    name: good.name,
    href: `/good/${good.id}`,
    imageSrc: PLACEHOLDER_IMAGE,
    imageAlt: good.name,
    price: {
      current: good.price,
      currency: 'USD',
      locale: 'en-US'
    },
    description: good.description
      ? good.description.length > 120
        ? `${good.description.slice(0, 120)}…`
        : good.description
      : undefined,
    badges: isAdding ? ADDING_BADGE : undefined,
    onAddToCart: isAdding ? undefined : () => onAddToCart(good),
    cta: {
      onQuickView: () => onQuickView(good),
      rating: 4.5,
      reviewCount: 24
    }
  };
}

export function ShopProductGrid({
  goods,
  title,
  className,
  gridClassName
}: {
  goods: Good[];
  title?: string;
  className?: string;
  gridClassName?: string;
}) {
  const router = useRouter();
  const { addCartItem, updateCartItem } = useCart();
  const [quickViewGood, setQuickViewGood] = useState<Good | null>(null);
  const [isQuickViewAdding, setIsQuickViewAdding] = useState(false);
  const [pendingGoodIds, setPendingGoodIds] = useState<Record<string, boolean>>({});

  const setGoodPending = (goodId: string, isPending: boolean) => {
    setPendingGoodIds((prev) => {
      const next = { ...prev };
      if (isPending) {
        next[goodId] = true;
      } else {
        delete next[goodId];
      }
      return next;
    });
  };

  const addGoodToCart = async (good: Good): Promise<boolean> => {
    addCartItem(goodToOptimisticVariant(good), good);
    try {
      const result = await addItem(undefined, good.id);
      if (!result.ok) {
        updateCartItem(good.id, 'minus');
        toast.error(result.message);
        return false;
      }
      toast.success(`${good.name} added to cart`);
      return true;
    } catch {
      updateCartItem(good.id, 'minus');
      toast.error('Error adding item to cart');
      return false;
    }
  };

  const handleAddToCart = async (good: Good) => {
    if (pendingGoodIds[good.id]) return;
    setGoodPending(good.id, true);
    await addGoodToCart(good);
    setGoodPending(good.id, false);
  };

  const handleQuickViewAddToCart = async () => {
    if (!quickViewGood || isQuickViewAdding) return;
    setIsQuickViewAdding(true);
    const added = await addGoodToCart(quickViewGood);
    if (added) {
      setQuickViewGood(null);
    }
    setIsQuickViewAdding(false);
  };

  const products = goods.map((good) =>
    goodToProduct(good, handleAddToCart, setQuickViewGood, !!pendingGoodIds[good.id])
  );

  return (
    <>
      <ProductGrid
        className={clsx('shop-productgrid', 'shop-productgrid--with-add-to-cart', className)}
        gridClassName={clsx('gap-4 sm:gap-6 lg:gap-8', gridClassName)}
        spacingX="lg"
        spacingY="lg"
        products={products}
        title={title}
        onProductClick={(product) => router.push(product.href)}
      />
      <ProductQuickView
        className={clsx(
          'shop-productquickview',
          'shop-productquickview--high-rating',
          isQuickViewAdding && 'pointer-events-none opacity-80'
        )}
        open={!!quickViewGood}
        onClose={() => {
          if (!isQuickViewAdding) {
            setQuickViewGood(null);
          }
        }}
        product={
          quickViewGood
            ? goodToQuickViewProduct(quickViewGood)
            : { name: '', imageSrc: '', imageAlt: '', price: '' }
        }
        onAddToCart={handleQuickViewAddToCart}
      />
    </>
  );
}
