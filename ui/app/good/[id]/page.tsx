import type { Metadata } from 'next';
import { notFound } from 'next/navigation';

import { GridTileImage } from 'components/grid/tile';
import { Gallery } from 'components/good/gallery';
import { GoodProvider } from 'components/good/good-context';
import { GoodDescription } from 'components/good/good-description';
import { getGood, getGoodRecommendations } from 'lib/shopify';
import { Image } from 'lib/shopify/types';
import Link from 'next/link';
import { Suspense } from 'react';

// DOCS: https://nextjs.org/docs/app/api-reference/file-conventions/route-segment-config#dynamic
export const dynamic = 'force-dynamic'

export async function generateMetadata(props: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const params = await props.params;
  const id = Number(params.id);

  if (!Number.isFinite(id)) return notFound();

  const good = await getGood(id);

  if (!good) return notFound();

  return {
    title: good.name,
    description: good.description,
  };
}

export default async function GoodPage(props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const id = Number(params.id);

  if (!Number.isFinite(id)) return notFound();

  const good = await getGood(id);

  if (!good) return notFound();

  const goodJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: good.name,
    description: good.description,
    offers: {
      '@type': 'Offer',
      priceCurrency: 'USD',
      price: good.price,
    }
  };

  // Mock images for demo
  const images: Image[] = [
    { url: 'https://picsum.photos/600/600?random=1', altText: good.name, width: 600, height: 600 },
    { url: 'https://picsum.photos/600/600?random=2', altText: good.name, width: 600, height: 600 },
    { url: 'https://picsum.photos/600/600?random=3', altText: good.name, width: 600, height: 600 },
    { url: 'https://picsum.photos/600/600?random=4', altText: good.name, width: 600, height: 600 },
    { url: 'https://picsum.photos/600/600?random=5', altText: good.name, width: 600, height: 600 },
  ];

  return (
    <GoodProvider>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(goodJsonLd)
        }}
      />
      <div className="mx-auto max-w-screen-2xl px-4">
        <div
          className='flex flex-col rounded-lg border border-neutral-200 bg-white p-8 md:p-12 lg:flex-row lg:gap-8 dark:border-neutral-800 dark:bg-black'>
          <div className='h-full w-full basis-full lg:basis-4/6'>
            <Suspense
              fallback={
                <div className='relative aspect-square h-full max-h-[550px] w-full overflow-hidden' />
              }
            >
              <Gallery
                images={images.slice(0, 5).map((image: Image) => ({
                  src: image.url,
                  altText: image.altText,
                }))}
              />
            </Suspense>
          </div>

          <div className='basis-full lg:basis-2/6'>
            <Suspense fallback={null}>
              <GoodDescription good={good} />
            </Suspense>
          </div>
        </div>
        {/*<RelatedGoods id={good.id} />*/}
      </div>
    </GoodProvider>
  );
}

async function RelatedGoods({ id }: { id: number }) {
  const relatedGoods = await getGoodRecommendations(id)

  if (!relatedGoods.length) return null

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
