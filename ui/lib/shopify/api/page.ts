import { shopifyFetch } from '../fetch';
import type { Page, ShopifyPageOperation, ShopifyPagesOperation } from '../types';
import { getPageQuery, getPagesQuery } from '../queries/page';

export async function getPage(id: number): Promise<Page> {
  const res = await shopifyFetch<ShopifyPageOperation>({
    query: getPageQuery,
    cache: 'no-store',
    variables: { id }
  });

  return res.body.data.pageByHandle;
}

export async function getPages(): Promise<Page[]> {
  const res = await shopifyFetch<ShopifyPagesOperation>({
    query: getPagesQuery,
    cache: 'no-store',
  });

  return res.body?.data?.pages
    ? res.body.data.pages.edges.map((edge) => edge.node)
    : [];
}
