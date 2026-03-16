'use client';

import { ProductGrid, ProductQuickView } from '@shortlink-org/ui-kit';
import clsx from 'clsx';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { addItem } from 'components/cart/actions';
import { useCart } from 'components/cart/cart-context';
import { DEFAULT_OPTION } from 'lib/constants';
import { Good, GoodVariant } from 'lib/shopify/types';
import { getStorefrontArtwork, getStorefrontCategory } from 'lib/storefront-art';
import { toast } from 'sonner';
import type { ProductGridProduct, ProductQuickViewProduct } from '@shortlink-org/ui-kit';

const productCardSlotClassNames = {
  description: 'min-h-[2.5rem]',
  footer: 'gap-2',
  price: 'min-w-0',
  cta: 'shrink-0'
};

function placeholderImage(good: Good, variant = 0): string {
  return getStorefrontArtwork(good.name, good.id, {
    width: 720,
    height: 900,
    variant,
    eyebrow: 'shortlink goods',
    subtitle: getStorefrontCategory(good.name)
  });
}
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
    imageSrc: placeholderImage(good, 1),
    imageAlt: good.name,
    price: formatPrice(good.price),
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
    imageSrc: placeholderImage(good),
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
    badges: isAdding
      ? ADDING_BADGE
      : [
          {
            label: good.price >= 30 ? 'Signature drop' : 'Quick pick',
            tone: good.price >= 30 ? 'warning' : 'success'
          }
        ],
    onAddToCart: isAdding ? undefined : () => onAddToCart(good),
    cta: {
      onQuickView: () => onQuickView(good)
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
  const { addCartItem, updateCartItem, setCartId } = useCart();
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
      setCartId(result.cartId);
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
        gridClassName={clsx('gap-3 sm:gap-4 lg:gap-6', gridClassName)}
        productClassName="shop-productgrid__card"
        cardSlotClassNames={productCardSlotClassNames}
        spacingX="lg"
        spacingY="lg"
        products={products}
        title={title}
        onProductClick={(product) => router.push(product.href)}
      />
      <ProductQuickView
        className={clsx(isQuickViewAdding && 'pointer-events-none opacity-80')}
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
