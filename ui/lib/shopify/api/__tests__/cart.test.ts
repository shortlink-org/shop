import { beforeEach, describe, expect, it, vi } from 'vitest';

import { shopifyFetch } from '../../fetch';
import { addToCart } from '../cart';
import { addToCartMutation } from '../../mutations/cart';

vi.mock('../../fetch', () => ({
  shopifyFetch: vi.fn()
}));

describe('cart api headers', () => {
  const mockedShopifyFetch = vi.mocked(shopifyFetch);

  beforeEach(() => {
    mockedShopifyFetch.mockReset();
    mockedShopifyFetch.mockResolvedValue({
      status: 200,
      body: { data: { addItem: { _: true } } }
    });
  });

  it('forwards Authorization and X-User-ID for addToCart', async () => {
    await addToCart([{ goodId: '243fde0d-8120-40d2-b899-c6cd158692bb', quantity: 1 }], {
      authorization: 'Bearer token',
      userId: '550e8400-e29b-41d4-a716-446655440000'
    });

    expect(mockedShopifyFetch).toHaveBeenCalledTimes(1);
    expect(mockedShopifyFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        query: addToCartMutation,
        headers: {
          Authorization: 'Bearer token',
          'X-User-ID': '550e8400-e29b-41d4-a716-446655440000'
        }
      })
    );
  });
});
