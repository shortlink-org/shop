import { getCollection } from 'lib/shopify';
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
  return (
    <section>
      <p className="py-3 text-lg">{`No goods found in this collection`}</p>
    </section>
  );
}
