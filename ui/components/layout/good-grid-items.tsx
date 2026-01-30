import Grid from 'components/grid';
import { GridTileImage } from 'components/grid/tile';
import { Good } from 'lib/shopify/types';
import Link from 'next/link';

export default function GoodGridItems({ goods }: { goods: Good[] }) {
  return (
    <>
      {goods.map((good) => (
        <Grid.Item key={good.id} className="animate-fadeIn">
          <Link
            className="relative inline-block h-full w-full"
            href={`/good/${good.id}`}
            prefetch={true}
          >
            <GridTileImage
              alt={good.name}
              label={{
                title: good.name,
                amount: good.price,
              }}
              src="https://picsum.photos/400"
              fill
              sizes="(min-width: 768px) 33vw, (min-width: 640px) 50vw, 100vw"
            />
          </Link>
        </Grid.Item>
      ))}
    </>
  );
}
