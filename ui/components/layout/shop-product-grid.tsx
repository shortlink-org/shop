'use client';

import { ProductGrid, ProductQuickView } from '@shortlink-org/ui-kit';
import clsx from 'clsx';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { addItem } from 'components/cart/actions';
import { Good } from 'lib/shopify/types';
import type { ProductGridProduct, ProductQuickViewProduct } from '@shortlink-org/ui-kit';

const PLACEHOLDER_IMAGE = 'https://picsum.photos/400/500';

function formatPrice(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
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
  onAddToCart: (goodId: string) => void,
  onQuickView: (good: Good) => void
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
    description: good.description ? (good.description.length > 120 ? `${good.description.slice(0, 120)}…` : good.description) : undefined,
    onAddToCart: () => onAddToCart(good.id),
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
  const [quickViewGood, setQuickViewGood] = useState<Good | null>(null);

  const handleAddToCart = async (goodId: string) => {
    await addItem(undefined, goodId);
  };

  const handleQuickViewAddToCart = async () => {
    if (!quickViewGood) return;
    await addItem(undefined, quickViewGood.id);
    setQuickViewGood(null);
  };

  const products: ProductGridProduct[] = goods.map((good) =>
    goodToProduct(good, handleAddToCart, setQuickViewGood)
  );

  return (
    <>
      <ProductGrid
        className={clsx('shop-productgrid', 'shop-productgrid--with-add-to-cart', className)}
        gridClassName={gridClassName}
        products={products}
        title={title}
        onProductClick={(product) => router.push(product.href)}
      />
      <ProductQuickView
        className="shop-productquickview shop-productquickview--high-rating"
        open={!!quickViewGood}
        onClose={() => setQuickViewGood(null)}
        product={quickViewGood ? goodToQuickViewProduct(quickViewGood) : { name: '', imageSrc: '', imageAlt: '', price: '' }}
        onAddToCart={handleQuickViewAddToCart}
      />
    </>
  );
}
