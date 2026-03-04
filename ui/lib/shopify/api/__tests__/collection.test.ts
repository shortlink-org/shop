import { beforeEach, describe, expect, it, vi } from 'vitest';

import { GOODS_UNAVAILABLE } from '../../sentinels';
import { shopifyFetch } from '../../fetch';
import { getCollectionProducts } from '../collection';
import {
  getCollectionProductsPage1Query,
  getCollectionProductsQuery
} from '../../queries/collection';

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

  it('normalizes non-integer page input and requests page=1', async () => {
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

    const result = await getCollectionProducts({
      page: '""243fde0d-8120-40d2-b899-c6cd158692bb""'
    });

    expect(result).not.toBe(GOODS_UNAVAILABLE);
    expect(mockedShopifyFetch).toHaveBeenCalledTimes(1);
    expect(mockedShopifyFetch).toHaveBeenCalledWith(
      expect.objectContaining({
        query: getCollectionProductsQuery,
        variables: { page: 1 }
      })
    );
  });

  it('falls back to literal page=1 query on Int type mismatch error', async () => {
    mockedShopifyFetch
      .mockRejectedValueOnce({
        error: {
          message: 'Int cannot represent non-integer value: ""243fde0d-8120-40d2-b899-c6cd158692bb""'
        }
      })
      .mockResolvedValueOnce({
        status: 200,
        body: {
          data: {
            goods: {
              results: [
                {
                  id: '2',
                  name: 'Recovered good',
                  price: 200,
                  description: 'desc',
                  created_at: '2026-03-01T00:00:00Z',
                  updated_at: '2026-03-01T00:00:00Z'
                }
              ]
            }
          }
        } as MockGoodsResponse
      });

    const result = await getCollectionProducts({
      page: '""243fde0d-8120-40d2-b899-c6cd158692bb""'
    });

    expect(mockedShopifyFetch).toHaveBeenCalledTimes(2);
    expect(mockedShopifyFetch).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        query: getCollectionProductsQuery,
        variables: { page: 1 }
      })
    );
    expect(mockedShopifyFetch).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        query: getCollectionProductsPage1Query
      })
    );
    expect(result).not.toBe(GOODS_UNAVAILABLE);
    expect(Array.isArray(result)).toBe(true);
    if (Array.isArray(result)) {
      expect(result[0]?.id).toBe('2');
    }
  });

  it('returns GOODS_UNAVAILABLE on non-Int errors', async () => {
    mockedShopifyFetch.mockRejectedValueOnce({
      error: { message: 'Failed to fetch from Subgraph' }
    });

    const result = await getCollectionProducts({ page: 1 });

    expect(mockedShopifyFetch).toHaveBeenCalledTimes(1);
    expect(result).toBe(GOODS_UNAVAILABLE);
  });
});
