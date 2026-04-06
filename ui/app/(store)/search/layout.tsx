import Collections from 'components/layout/search/collections';
import FilterList from 'components/layout/search/filter';
import { sorting } from 'lib/constants';

// DOCS: https://nextjs.org/docs/app/api-reference/file-conventions/route-segment-config#dynamic
export const dynamic = 'force-dynamic';

export default function SearchLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-6 pb-4 text-black md:flex-row md:gap-8 dark:text-white">
      <div className="order-first w-full flex-none md:max-w-[10rem] md:pt-2">
        <Collections />
      </div>
      <div className="order-last min-h-screen min-w-0 flex-1 md:order-none">{children}</div>
      <div className="order-none flex-none md:order-last md:w-[7.5rem] md:pt-2">
        <FilterList list={sorting} title="Sort by" />
      </div>
    </div>
  );
}
