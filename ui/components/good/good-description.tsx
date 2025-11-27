import { AddToCart } from 'components/cart/add-to-cart';
import Price from 'components/price';
import Prose from 'components/prose';
import { Good } from 'lib/shopify/types';
import { VariantSelector } from './variant-selector';
import { GoodOption, GoodVariant } from 'lib/shopify/types';

export function GoodDescription({ good }: { good: Good }) {
  return (
    <>
      <div className="mb-6 flex flex-col border-b pb-6 dark:border-neutral-700">
        <h1 className="mb-2 text-5xl font-medium">{good.name}</h1>
        <div className="mr-auto w-auto rounded-full bg-blue-600 p-2 text-sm text-white">
          <Price
            amount={good.price}
          />
        </div>
      </div>
      {/*<VariantSelector options={good.options} variants={good.variants} />*/}
      {/*{good.descriptionHtml ? (*/}
      {/*  <Prose*/}
      {/*    className="mb-6 text-sm leading-tight dark:text-white/[60%]"*/}
      {/*    html={good.descriptionHtml}*/}
      {/*  />*/}
      {/*) : null}*/}
      <AddToCart good={good} />
    </>
  );
}

