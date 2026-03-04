import { ShopProductGrid } from 'components/layout/shop-product-grid';
import { RetryButton } from 'components/retry-button';
import { getCollectionProducts, GOODS_UNAVAILABLE } from 'lib/shopify';
import { headers } from 'next/headers';

export const metadata = {
  description: 'High-performance ecommerce store built with Next.js, Vercel, and Shopify.',
  openGraph: {
    type: 'website'
  }
};

export default async function HomePage(_props: {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const authHeader = (await headers()).get('authorization') ?? undefined;
  // Never pass searchParams.page (or any URL param) into getCollectionProducts — BFF expects Int.
  // If we ever add pagination, parse page from searchParams and pass only a normalized integer.
  const homepageItems = await getCollectionProducts({}, { authorization: authHeader });

  if (homepageItems === GOODS_UNAVAILABLE) {
    return (
      <section className="mx-auto max-w-screen-2xl px-4 pb-4">
        <div className="flex flex-col items-center justify-center rounded-lg border border-neutral-200 bg-neutral-50 py-16 dark:border-neutral-800 dark:bg-neutral-900">
          <p className="text-lg font-semibold">We couldn&apos;t load products</p>
          <p className="mt-2 text-center text-sm text-neutral-500 dark:text-neutral-400">
            We&apos;ll show them when they&apos;re available again.
          </p>
          <RetryButton />
        </div>
      </section>
    );
  }

  if (homepageItems.length === 0) return null;

  return (
    <section className="mx-auto max-w-screen-2xl px-4 pb-4">
      <ShopProductGrid goods={homepageItems} />
    </section>
  );
}
