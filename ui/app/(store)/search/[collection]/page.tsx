import { getCollection, GOODS_UNAVAILABLE } from 'lib/shopify';
import { Metadata } from 'next';
import { headers } from 'next/headers';
import { notFound } from 'next/navigation';

export async function generateMetadata(props: {
  params: Promise<{ collection: string }>;
}): Promise<Metadata> {
  const params = await props.params;
  const authHeader = (await headers()).get('authorization') ?? undefined;
  const collection = await getCollection(Number(params.collection), {
    authorization: authHeader
  });

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
  const authHeader = (await headers()).get('authorization') ?? undefined;
  const collection = await getCollection(Number(params.collection), {
    authorization: authHeader
  });

  if (collection === GOODS_UNAVAILABLE) {
    return (
      <section className="px-4 py-8">
        <div className="flex flex-col items-center justify-center rounded-2xl border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_94%,transparent)] py-16 shadow-[0_16px_40px_-28px_rgba(15,23,42,0.18)] dark:shadow-[0_16px_48px_-28px_rgba(0,0,0,0.45)]">
          <p className="text-lg font-semibold text-[var(--color-foreground)]">We couldn&apos;t load this collection</p>
          <p className="mt-2 max-w-md text-center text-sm text-[var(--color-muted-foreground)]">
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
