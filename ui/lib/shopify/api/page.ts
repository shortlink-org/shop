import { shopifyFetch } from '../fetch';
import type { Page, ShopifyPageOperation, ShopifyPagesOperation } from '../types';
import { getPageQuery, getPagesQuery } from '../queries/page';

export type RequestOptions = { authorization?: string };

export async function getPage(id: number, options?: RequestOptions): Promise<Page> {
  const res = await shopifyFetch<ShopifyPageOperation>({
    query: getPageQuery,
    cache: 'no-store',
    variables: { id },
    headers: options?.authorization ? { Authorization: options.authorization } : {}
  });

  return res.body.data.pageByHandle;
}

export async function getPages(options?: RequestOptions): Promise<Page[]> {
  const res = await shopifyFetch<ShopifyPagesOperation>({
    query: getPagesQuery,
    cache: 'no-store',
    headers: options?.authorization ? { Authorization: options.authorization } : {}
  });

  return res.body?.data?.pages
    ? res.body.data.pages.edges.map((edge) => edge.node)
    : [];
}
