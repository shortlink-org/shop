import Grid from 'components/grid';
import { GridTileImage } from 'components/grid/tile';
import { Good } from 'lib/shopify/types';
import Link from 'next/link';

export default function GoodGridItems({ goods }: { goods: Good[] }) {
  return (
    <>
      {goods.map((good) => (
        <Grid.Item key={good.handle} className="animate-fadeIn">
          <Link
            className="relative inline-block h-full w-full"
            href={`/good/${good.handle}`}
            prefetch={true}
          >
            <GridTileImage
              alt={good.title}
              label={{
                title: good.title,
                amount: good.priceRange.maxVariantPrice.amount,
                currencyCode: good.priceRange.maxVariantPrice.currencyCode
              }}
              src={good.featuredImage?.url}
              fill
              sizes="(min-width: 768px) 33vw, (min-width: 640px) 50vw, 100vw"
            />
          </Link>
        </Grid.Item>
      ))}
    </>
  );
}
