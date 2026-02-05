import Grid from 'components/grid';
import GoodGridItems from 'components/layout/good-grid-items';
import { RetryButton } from 'components/retry-button';
import { defaultSort, sorting } from 'lib/constants';
import { getGoods, GOODS_UNAVAILABLE } from 'lib/shopify';

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

  const goods = await getGoods({ sortKey, reverse, query: searchValue });

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

  return (
    <>
      {searchValue ? (
        <p className="mb-4">
          {goods.length === 0
            ? 'There are no goods that match '
            : `Showing ${goods.length} ${resultsText} for `}
          <span className="font-bold">&quot;{searchValue}&quot;</span>
        </p>
      ) : null}
      {goods.length > 0 ? (
        <Grid className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
          <GoodGridItems goods={goods} />
        </Grid>
      ) : null}
    </>
  );
}
