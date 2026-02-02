import { getCollection, GOODS_UNAVAILABLE } from 'lib/shopify';
import { Metadata } from 'next';
import { notFound } from 'next/navigation';

import Grid from 'components/grid';
import GoodGridItems from 'components/layout/good-grid-items';
import { defaultSort, sorting } from 'lib/constants';

export async function generateMetadata(props: {
  params: Promise<{ collection: string }>;
}): Promise<Metadata> {
  const params = await props.params;
  const collection = await getCollection(Number(params.collection));

  if (collection === GOODS_UNAVAILABLE) {
    return { title: 'Collection unavailable' };
  }
  if (!collection) return notFound();

  return {
    title: collection.seo?.title || collection.title,
    description:
      collection.seo?.description || collection.description || `${collection.title} goods`
  };
}

export default async function CategoryPage(props: {
  params: Promise<{ collection: string }>;
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const params = await props.params;
  const collection = await getCollection(Number(params.collection));

  if (collection === GOODS_UNAVAILABLE) {
    return (
      <section className="px-4 py-8">
        <div className="flex flex-col items-center justify-center rounded-lg border border-neutral-200 bg-neutral-50 py-16 dark:border-neutral-800 dark:bg-neutral-900">
          <p className="text-lg font-semibold">We couldn&apos;t load this collection</p>
          <p className="mt-2 text-center text-sm text-neutral-500 dark:text-neutral-400">
            We&apos;ll show it when it&apos;s available again.
          </p>
        </div>
      </section>
    );
  }
  if (!collection) return notFound();

  return (
    <section>
      <p className="py-3 text-lg">{`No goods found in this collection`}</p>
    </section>
  );
}
