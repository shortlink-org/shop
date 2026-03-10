import { ShopProductGrid } from 'components/layout/shop-product-grid';
import { RetryButton } from 'components/retry-button';
import { defaultSort, sorting } from 'lib/constants';
import { getGoods, GOODS_UNAVAILABLE } from 'lib/shopify';
import { getBaseUrl, sanitizeJsonLd } from 'lib/utils';
import { headers } from 'next/headers';

export const metadata = {
  title: 'Search',
  description: 'Search for products in the store.'
};

export default async function SearchPage(props: {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const searchParams = await props.searchParams;
  const { sort, q: searchValue } = (searchParams ?? {}) as { [key: string]: string };
  const { sortKey, reverse } = sorting.find((item) => item.slug === sort) || defaultSort;
  const authHeader = (await headers()).get('authorization') ?? undefined;

  const goods = await getGoods(
    { sortKey, reverse, query: searchValue },
    { authorization: authHeader }
  );

  if (goods === GOODS_UNAVAILABLE) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-neutral-200 bg-neutral-50 py-16 dark:border-neutral-800 dark:bg-neutral-900">
        <p className="text-lg font-semibold">We couldn&apos;t load products</p>
        <p className="mt-2 text-center text-sm text-neutral-500 dark:text-neutral-400">
          We&apos;ll show them when they&apos;re available again.
        </p>
        <RetryButton />
      </div>
    );
  }

  const resultsText = goods.length > 1 ? 'results' : 'result';

  const itemListJsonLd =
    goods.length > 0
      ? {
          '@context': 'https://schema.org',
          '@type': 'ItemList',
          numberOfItems: goods.length,
          itemListElement: goods.map((good, index) => ({
            '@type': 'ListItem',
            position: index + 1,
            url: `${getBaseUrl()}/good/${good.id}`,
            name: good.name
          }))
        }
      : null;

  return (
    <>
      {itemListJsonLd ? (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: sanitizeJsonLd(itemListJsonLd) }}
        />
      ) : null}
      {searchValue ? (
        <p className="mb-4">
          {goods.length === 0
            ? 'No products found for '
            : `Showing ${goods.length} ${resultsText} for `}
          <span className="font-bold">&quot;{searchValue}&quot;</span>
        </p>
      ) : null}
      {goods.length > 0 ? <ShopProductGrid goods={goods} /> : null}
    </>
  );
}
