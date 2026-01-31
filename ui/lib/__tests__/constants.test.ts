import { describe, expect, it } from 'vitest';
import {
  defaultSort,
  sorting,
  TAGS,
  HIDDEN_GOOD_TAG,
  DEFAULT_OPTION,
  SHOPIFY_GRAPHQL_API_ENDPOINT,
  SortFilterItem
} from '../constants';

describe('defaultSort', () => {
  it('should have correct default values', () => {
    expect(defaultSort.title).toBe('Relevance');
    expect(defaultSort.slug).toBeNull();
    expect(defaultSort.sortKey).toBe('RELEVANCE');
    expect(defaultSort.reverse).toBe(false);
  });
});

describe('sorting', () => {
  it('should contain all sorting options', () => {
    expect(sorting).toHaveLength(5);
  });

  it('should have defaultSort as first option', () => {
    expect(sorting[0]).toBe(defaultSort);
  });

  it('should have correct sorting options', () => {
    const titles = sorting.map((s) => s.title);
    expect(titles).toContain('Relevance');
    expect(titles).toContain('Trending');
    expect(titles).toContain('Latest arrivals');
    expect(titles).toContain('Price: Low to high');
    expect(titles).toContain('Price: High to low');
  });

  it('should have valid sortKeys', () => {
    const validSortKeys = ['RELEVANCE', 'BEST_SELLING', 'CREATED_AT', 'PRICE'];
    sorting.forEach((item) => {
      expect(validSortKeys).toContain(item.sortKey);
    });
  });

  it('should have unique slugs (except null)', () => {
    const slugs = sorting.filter((s) => s.slug !== null).map((s) => s.slug);
    const uniqueSlugs = new Set(slugs);
    expect(slugs.length).toBe(uniqueSlugs.size);
  });

  it('should have correct price sorting configuration', () => {
    const priceLowToHigh = sorting.find((s) => s.slug === 'price-asc');
    const priceHighToLow = sorting.find((s) => s.slug === 'price-desc');

    expect(priceLowToHigh).toBeDefined();
    expect(priceLowToHigh?.sortKey).toBe('PRICE');
    expect(priceLowToHigh?.reverse).toBe(false);

    expect(priceHighToLow).toBeDefined();
    expect(priceHighToLow?.sortKey).toBe('PRICE');
    expect(priceHighToLow?.reverse).toBe(true);
  });
});

describe('TAGS', () => {
  it('should have all required tag keys', () => {
    expect(TAGS.collections).toBe('collections');
    expect(TAGS.goods).toBe('goods');
    expect(TAGS.cart).toBe('cart');
  });

  it('should have exactly 3 tags', () => {
    expect(Object.keys(TAGS)).toHaveLength(3);
  });
});

describe('constants', () => {
  it('should have correct HIDDEN_GOOD_TAG', () => {
    expect(HIDDEN_GOOD_TAG).toBe('nextjs-frontend-hidden');
  });

  it('should have correct DEFAULT_OPTION', () => {
    expect(DEFAULT_OPTION).toBe('Default Title');
  });

  it('should have correct SHOPIFY_GRAPHQL_API_ENDPOINT', () => {
    expect(SHOPIFY_GRAPHQL_API_ENDPOINT).toBe('/api/2023-01/graphql.json');
  });
});

describe('SortFilterItem type', () => {
  it('should accept valid SortFilterItem objects', () => {
    const item: SortFilterItem = {
      title: 'Test',
      slug: 'test',
      sortKey: 'PRICE',
      reverse: false
    };
    expect(item).toBeDefined();
  });

  it('should accept null slug', () => {
    const item: SortFilterItem = {
      title: 'Test',
      slug: null,
      sortKey: 'RELEVANCE',
      reverse: false
    };
    expect(item.slug).toBeNull();
  });
});
