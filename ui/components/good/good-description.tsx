'use client';

import { ProductDescription } from '@/lib/ui-kit';
import Price from 'components/price';
import { Good } from 'lib/shopify/types';

export function GoodDescription({ good }: { good: Good }) {
  return (
    <section className="shop-productdescription">
      <div className="shop-productdescription__header mb-6 flex flex-col border-b border-neutral-200 pb-6 dark:border-neutral-700">
        <h1 className="mb-2 text-3xl font-semibold tracking-tight text-neutral-900 md:text-4xl lg:text-5xl dark:text-neutral-100">
          {good.name}
        </h1>
        <div className="shop-productdescription__price mr-auto w-auto rounded-full bg-blue-600 px-3 py-2 text-sm font-medium text-white">
          <Price amount={good.price} />
        </div>
      </div>
      <ProductDescription
        className="shop-productdescription__content"
        description={good.description ?? undefined}
        highlights={[]}
        details={undefined}
      />
    </section>
  );
}
