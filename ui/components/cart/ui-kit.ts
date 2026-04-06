import type { BasketItem as BasketItemData } from '@/lib/ui-kit';
import { DEFAULT_OPTION } from 'lib/constants';
import type { Cart, CartItem } from 'lib/shopify/types';
import { getStorefrontArtwork, getStorefrontCategory } from 'lib/storefront-art';

export const CART_ITEM_PLACEHOLDER_IMAGE = getStorefrontArtwork('Shortlink goods', 'cart', {
  width: 320,
  height: 320,
  eyebrow: 'cart preview',
  subtitle: 'branded selection'
});

export function formatCartMoney(amount: number, currencyCode = 'USD', locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: currencyCode
  }).format(amount);
}

function getUnitPrice(item: CartItem): number {
  if (item.quantity <= 0) {
    return 0;
  }

  return Number(item.cost.totalAmount.amount) / item.quantity;
}

export function cartItemToBasketItem(item: CartItem): BasketItemData {
  const productImage = item.merchandise.product.featuredImage;
  const variantTitle =
    item.merchandise.title !== DEFAULT_OPTION ? item.merchandise.title : undefined;

  return {
    id: item.merchandise.id,
    name: item.merchandise.product.title,
    href: `/good/${item.merchandise.product.id}`,
    color: variantTitle,
    price: formatCartMoney(getUnitPrice(item), item.cost.totalAmount.currencyCode),
    quantity: item.quantity,
    imageSrc:
      productImage?.url ??
      getStorefrontArtwork(item.merchandise.product.title, item.merchandise.product.id, {
        width: 320,
        height: 320,
        eyebrow: 'cart preview',
        subtitle: getStorefrontCategory(item.merchandise.product.title)
      }),
    imageAlt: productImage?.altText || `${item.merchandise.product.title} image`
  };
}

export function cartToBasketItems(cart: Cart | undefined): BasketItemData[] {
  return [...(cart?.lines ?? [])]
    .sort((left, right) =>
      left.merchandise.product.title.localeCompare(right.merchandise.product.title)
    )
    .map(cartItemToBasketItem);
}
