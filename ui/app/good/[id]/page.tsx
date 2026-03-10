import type { Metadata } from 'next';
import { headers } from 'next/headers';
import { notFound } from 'next/navigation';

import { GridTileImage } from 'components/grid/tile';
import { GoodProductPage } from 'components/good/good-product-page';
import { GoodProvider } from 'components/good/good-context';
import { getGood, getGoodRecommendations, GOODS_UNAVAILABLE } from 'lib/shopify';
import { Image } from 'lib/shopify/types';
import { getBaseUrl, sanitizeJsonLd } from 'lib/utils';
import Link from 'next/link';
import { Suspense } from 'react';

// DOCS: https://nextjs.org/docs/app/api-reference/file-conventions/route-segment-config#dynamic
export const dynamic = 'force-dynamic'

export async function generateMetadata(props: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const params = await props.params;
  const id = params.id?.trim() ?? '';
  const authHeader = (await headers()).get('authorization') ?? undefined;

  if (!id) return notFound();

  const good = await getGood(id, { authorization: authHeader });

  if (good === GOODS_UNAVAILABLE) {
    return { title: 'Product unavailable' };
  }
  if (!good) return notFound();

  return {
    title: good.name,
    description: good.description,
  };
}

export default async function GoodPage(props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const id = params.id?.trim() ?? '';
  const authHeader = (await headers()).get('authorization') ?? undefined;

  if (!id) return notFound();

  const good = await getGood(id, { authorization: authHeader });

  if (good === GOODS_UNAVAILABLE) {
    return (
      <div className="mx-auto max-w-screen-2xl px-4 py-16">
        <div className="flex flex-col items-center justify-center rounded-lg border border-neutral-200 bg-neutral-50 py-16 dark:border-neutral-800 dark:bg-neutral-900">
          <p className="text-lg font-semibold">We couldn&apos;t load this product</p>
          <p className="mt-2 text-center text-sm text-neutral-500 dark:text-neutral-400">
            We&apos;ll show it when it&apos;s available again.
          </p>
        </div>
      </div>
    );
  }
  if (!good) return notFound();

  // Mock images for demo (stable per product via seed)
  const images: Image[] = [
    { url: `https://picsum.photos/seed/${encodeURIComponent(id)}-1/600/600`, altText: good.name, width: 600, height: 600 },
    { url: `https://picsum.photos/seed/${encodeURIComponent(id)}-2/600/600`, altText: good.name, width: 600, height: 600 },
    { url: `https://picsum.photos/seed/${encodeURIComponent(id)}-3/600/600`, altText: good.name, width: 600, height: 600 },
    { url: `https://picsum.photos/seed/${encodeURIComponent(id)}-4/600/600`, altText: good.name, width: 600, height: 600 },
    { url: `https://picsum.photos/seed/${encodeURIComponent(id)}-5/600/600`, altText: good.name, width: 600, height: 600 },
  ];

  const galleryImages = images.slice(0, 5).map((image: Image) => ({
    src: image.url,
    altText: image.altText
  }));

  const goodJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: good.name,
    description: good.description,
    image: images.slice(0, 3).map((img) => img.url),
    url: `${getBaseUrl()}/good/${good.id}`,
    offers: {
      '@type': 'Offer',
      priceCurrency: 'USD',
      price: good.price,
    },
  };

  return (
    <GoodProvider>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: sanitizeJsonLd(goodJsonLd),
        }}
      />
      <div className="mx-auto max-w-screen-2xl px-4 pb-16">
        <Suspense fallback={null}>
          <GoodProductPage good={good} images={galleryImages} />
        </Suspense>
        {/*<RelatedGoods id={good.id} />*/}
      </div>
    </GoodProvider>
  );
}

async function RelatedGoods({ id }: { id: string }) {
  const authHeader = (await headers()).get('authorization') ?? undefined;
  const relatedGoods = await getGoodRecommendations(id, {
    authorization: authHeader
  });

  if (relatedGoods === GOODS_UNAVAILABLE || !relatedGoods.length) return null;

  return (
    <div className='py-8'>
      <h2 className='mb-4 text-2xl font-bold'>Related Goods</h2>
      <ul className="flex w-full gap-4 overflow-x-auto pt-1">
        {relatedGoods.map((good) => (
          <li
            key={good.id}
            className="aspect-square w-full flex-none min-[475px]:w-1/2 sm:w-1/3 md:w-1/4 lg:w-1/5"
          >
            <Link
              className="relative h-full w-full"
              href={`/good/${good.id}`}
              prefetch={true}
            >
              <GridTileImage
                alt={good.name}
                label={{
                  title: good.name,
                  amount: good.price,
                }}
                src={"https://picsum.photos/200"}
                fill
                sizes="(min-width: 1024px) 20vw, (min-width: 768px) 25vw, (min-width: 640px) 33vw, (min-width: 475px) 50vw, 100vw"
              />
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
