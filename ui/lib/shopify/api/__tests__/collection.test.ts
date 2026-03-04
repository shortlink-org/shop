import { beforeEach, describe, expect, it, vi } from 'vitest';

import { GOODS_UNAVAILABLE } from '../../sentinels';
import { shopifyFetch } from '../../fetch';
import { getCollectionProducts } from '../collection';
import { getCollectionProductsQuery } from '../../queries/collection';

vi.mock('../../fetch', () => ({
  shopifyFetch: vi.fn()
}));

type MockGoodsResponse = {
  data: {
    goods: {
      results: Array<{
        id: string;
        name: string;
        price: number;
        description: string;
        created_at: string;
        updated_at: string;
      }>;
    };
  };
};

describe('getCollectionProducts', () => {
  const mockedShopifyFetch = vi.mocked(shopifyFetch);

  beforeEach(() => {
    mockedShopifyFetch.mockReset();
  });

  it('requests goods without page variables', async () => {
    mockedShopifyFetch.mockResolvedValueOnce({
      status: 200,
      body: {
        data: {
          goods: {
            results: [
              {
                id: '1',
                name: 'Test good',
                price: 100,
                description: 'desc',
                created_at: '2026-03-01T00:00:00Z',
                updated_at: '2026-03-01T00:00:00Z'
              }
            ]
          }
        }
      } as MockGoodsResponse
    });

    const result = await getCollectionProducts();

    expect(result).not.toBe(GOODS_UNAVAILABLE);
    expect(mockedShopifyFetch).toHaveBeenCalledTimes(1);
    expect(mockedShopifyFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        query: getCollectionProductsQuery
      })
    );
    expect(mockedShopifyFetch.mock.calls[0]?.[0]).not.toHaveProperty('variables');
  });

  it('returns GOODS_UNAVAILABLE on GraphQL errors', async () => {
    mockedShopifyFetch.mockRejectedValueOnce({
      error: {
        message: 'Int cannot represent non-integer value: ""243fde0d-8120-40d2-b899-c6cd158692bb""'
      }
    });

    const result = await getCollectionProducts();

    expect(mockedShopifyFetch).toHaveBeenCalledTimes(1);
    expect(mockedShopifyFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        query: getCollectionProductsQuery
      })
    );
    expect(mockedShopifyFetch.mock.calls[0]?.[0]).not.toHaveProperty('variables');
    expect(result).toBe(GOODS_UNAVAILABLE);
  });

  it('returns GOODS_UNAVAILABLE on non-Int errors', async () => {
    mockedShopifyFetch.mockRejectedValueOnce({
      error: { message: 'Failed to fetch from Subgraph' }
    });

    const result = await getCollectionProducts();

    expect(mockedShopifyFetch).toHaveBeenCalledTimes(1);
    expect(result).toBe(GOODS_UNAVAILABLE);
  });
});
