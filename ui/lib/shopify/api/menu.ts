import { TAGS } from 'lib/constants';
import { domain } from '../config';
import { shopifyFetch } from '../fetch';
import type { Menu, ShopifyMenuOperation } from '../types';
import { getMenuQuery } from '../queries/menu';

export async function getMenu(id: number): Promise<Menu[]> {
  const res = await shopifyFetch<ShopifyMenuOperation>({
    query: getMenuQuery,
    tags: [TAGS.collections],
    variables: {
      id
    }
  });

  return (
    res.body?.data?.menu?.items.map((item: { title: string; url: string }) => ({
      title: item.title,
      path: item.url.replace(domain, '').replace('/collections', '/search').replace('/pages', '')
    })) || []
  );
}
