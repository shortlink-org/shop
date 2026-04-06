'use client';

import { Breadcrumbs, ProductPage } from '@/lib/ui-kit';
import { AddToCartBlock } from 'components/good/add-to-cart-block';
import { BackButton } from 'components/good/back-button';
import { Gallery } from 'components/good/gallery';
import { Good } from 'lib/shopify/types';

const baseBreadcrumbs = [
  { id: 'home', name: 'Home', href: '/' },
  { id: 'search', name: 'Search', href: '/search' }
];

function formatPrice(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount);
}

export function GoodProductPage({
  good,
  images
}: {
  good: Good;
  images: { src: string; altText: string }[];
}) {
  const galleryImages = images.slice(0, 5);

  return (
    <ProductPage
      className="shop-productpage"
      name={good.name}
      price={formatPrice(good.price)}
      href={`/good/${good.id}`}
      breadcrumbs={baseBreadcrumbs}
      images={galleryImages.map((img) => ({ src: img.src, alt: img.altText }))}
      colors={[]}
      sizes={[]}
      description={good.description ?? undefined}
      headerSlot={
        <nav
          className="mb-4 flex flex-row flex-wrap items-center gap-2 sm:gap-4"
          aria-label="Breadcrumb and back"
        >
          <BackButton goodId={good.id} className="shrink-0" />
          <span className="hidden shrink-0 sm:inline" aria-hidden>
            <span className="text-neutral-400 dark:text-neutral-500">|</span>
          </span>
          <div className="min-w-0 flex-1">
            <Breadcrumbs
              breadcrumbs={[
                ...baseBreadcrumbs,
                { id: 'product', name: good.name, href: `/good/${good.id}` }
              ]}
            />
          </div>
        </nav>
      }
      gallerySlot={<Gallery images={galleryImages} />}
      actionSlot={<AddToCartBlock good={good} />}
    />
  );
}
