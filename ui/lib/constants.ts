export type SortFilterItem = {
  title: string;
  slug: string | null;
  sortKey: 'RELEVANCE' | 'BEST_SELLING' | 'CREATED_AT' | 'PRICE';
  reverse: boolean;
};

export const defaultSort: SortFilterItem = {
  title: 'Relevance',
  slug: null,
  sortKey: 'RELEVANCE',
  reverse: false
};

export const SORT_SLUGS = {
  trending: 'trending-desc',
  latest: 'latest-desc',
  priceAsc: 'price-asc',
  priceDesc: 'price-desc'
} as const;

export const sorting: SortFilterItem[] = [
  defaultSort,
  { title: 'Trending', slug: SORT_SLUGS.trending, sortKey: 'BEST_SELLING', reverse: false },
  { title: 'Latest arrivals', slug: SORT_SLUGS.latest, sortKey: 'CREATED_AT', reverse: true },
  { title: 'Price: Low to high', slug: SORT_SLUGS.priceAsc, sortKey: 'PRICE', reverse: false },
  { title: 'Price: High to low', slug: SORT_SLUGS.priceDesc, sortKey: 'PRICE', reverse: true }
];

export const HTTP_STATUS_RATE_LIMIT = 429;

export const RATE_LIMIT_MESSAGE =
  'Too many requests. Please try again in a moment.';

export const TAGS = {
  collections: 'collections',
  goods: 'goods',
  cart: 'cart'
};

export const HIDDEN_GOOD_TAG = 'nextjs-frontend-hidden';
export const DEFAULT_OPTION = 'Default Title';
export const SHOPIFY_GRAPHQL_API_ENDPOINT = '/api/2023-01/graphql.json';
