import { describe, expect, it } from 'vitest';
import { DEFAULT_OPTION } from 'lib/constants';
import type { Cart, CartItem } from 'lib/shopify/types';
import { cartItemToBasketItem, cartToBasketItems, formatCartMoney } from '../ui-kit';

function createCartItem(overrides: Partial<CartItem> = {}): CartItem {
  return {
    id: 'line-1',
    quantity: 2,
    cost: {
      totalAmount: {
        amount: 50,
        currencyCode: 'USD'
      }
    },
    merchandise: {
      id: 'variant-1',
      title: DEFAULT_OPTION,
      selectedOptions: [],
      product: {
        id: 'product-1',
        handle: 'product-1',
        title: 'Alpha',
        featuredImage: {
          url: 'https://example.com/image.jpg',
          altText: 'Alpha image',
          width: 120,
          height: 120
        }
      }
    },
    ...overrides
  };
}

describe('cart ui-kit helpers', () => {
  it('formats money with the provided currency code', () => {
    expect(formatCartMoney(99.5, 'EUR')).toContain('99');
  });

  it('maps cart item data to basket item data', () => {
    const basketItem = cartItemToBasketItem(
      createCartItem({
        merchandise: {
          id: 'variant-42',
          title: 'Blue / XL',
          selectedOptions: [],
          product: {
            id: 'product-42',
            handle: 'product-42',
            title: 'Weekend Hoodie',
            featuredImage: {
              url: 'https://example.com/hoodie.jpg',
              altText: 'Weekend Hoodie image',
              width: 120,
              height: 120
            }
          }
        }
      })
    );

    expect(basketItem).toMatchObject({
      id: 'variant-42',
      name: 'Weekend Hoodie',
      href: '/good/product-42',
      color: 'Blue / XL',
      quantity: 2,
      imageSrc: 'https://example.com/hoodie.jpg',
      imageAlt: 'Weekend Hoodie image'
    });
    expect(basketItem.price).toContain('$25');
  });

  it('omits the default variant title from the secondary label', () => {
    const basketItem = cartItemToBasketItem(createCartItem());
    expect(basketItem.color).toBeUndefined();
  });

  it('sorts basket items by product title', () => {
    const cart: Cart = {
      id: 'cart-1',
      checkoutUrl: '',
      totalQuantity: 2,
      cost: {
        subtotalAmount: { amount: 50, currencyCode: 'USD' },
        totalAmount: { amount: 50, currencyCode: 'USD' },
        totalTaxAmount: { amount: 0, currencyCode: 'USD' }
      },
      lines: [
        createCartItem(),
        createCartItem({
          id: 'line-2',
          merchandise: {
            id: 'variant-2',
            title: DEFAULT_OPTION,
            selectedOptions: [],
            product: {
              id: 'product-2',
              handle: 'product-2',
              title: 'Aardvark Tee',
              featuredImage: {
                url: 'https://example.com/aardvark.jpg',
                altText: 'Aardvark Tee image',
                width: 120,
                height: 120
              }
            }
          }
        })
      ]
    };

    expect(cartToBasketItems(cart).map((item) => item.name)).toEqual(['Aardvark Tee', 'Alpha']);
  });
});
