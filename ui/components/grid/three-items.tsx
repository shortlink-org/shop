import { GridTileImage } from 'components/grid/tile';
import { getCollectionProducts } from 'lib/shopify';
import type { Good } from 'lib/shopify/types';
import Link from 'next/link';

function ThreeItemGridItem({
                             item,
                             size,
                             priority
                           }: {
  item: Good;
  size: 'full' | 'half';
  priority?: boolean;
}) {
  return (
    <div
      className={size === 'full' ? 'md:col-span-4 md:row-span-2' : 'md:col-span-2 md:row-span-1'}
    >
      <Link
        className="relative block aspect-square h-full w-full"
        href={`/good/${item.id}`}
        prefetch={true}
      >
        <GridTileImage
          src={"https://picsum.photos/200"}
          fill
          sizes={
            size === 'full' ? '(min-width: 768px) 66vw, 100vw' : '(min-width: 768px) 33vw, 100vw'
          }
          priority={priority}
          alt={item.name}
          label={{
            position: size === 'full' ? 'center' : 'bottom',
            title: item.name as string,
            amount: item.price,
            // currencyCode: item.priceRange?.maxVariantPrice?.currencyCode
          }}
        />
      </Link>
    </div>
  );
}

export async function ThreeItemGrid() {
  // Collections that start with `hidden-*` are hidden from the search page.
  const homepageItems = await getCollectionProducts({});

  if (!homepageItems[0] || !homepageItems[1] || !homepageItems[2]) return null;

  const [firstGood, secondGood, thirdGood] = homepageItems;

  return (
    <section className="mx-auto grid max-w-screen-2xl gap-4 px-4 pb-4 md:grid-cols-6 md:grid-rows-2 lg:max-h-[calc(100vh-200px)]">
      <ThreeItemGridItem size="full" item={firstGood} priority={true} />
      <ThreeItemGridItem size="half" item={secondGood} priority={true} />
      <ThreeItemGridItem size="half" item={thirdGood} />
    </section>
  );
}
