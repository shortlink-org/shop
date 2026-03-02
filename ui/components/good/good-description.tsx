'use client';

import { ProductDescription } from '@shortlink-org/ui-kit';
import Price from 'components/price';
import { Good } from 'lib/shopify/types';

export function GoodDescription({ good }: { good: Good }) {
  return (
    <>
      <div className="mb-6 flex flex-col border-b pb-6 dark:border-neutral-700">
        <h1 className="mb-2 text-5xl font-medium">{good.name}</h1>
        <div className="mr-auto w-auto rounded-full bg-blue-600 p-2 text-sm text-white">
          <Price amount={good.price} />
        </div>
      </div>
      <ProductDescription
        description={good.description ?? undefined}
        highlights={[]}
        details={undefined}
      />
    </>
  );
}
