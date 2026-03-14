import { getOrderTrackingPageQuery } from '../queries/order';
import { shopifyFetch } from '../fetch';
import type { OrderTrackingPageData, ShopifyOrderTrackingPageOperation } from '../types';

export type RequestOptions = { authorization?: string };

export async function getOrderTrackingPage(
  id: string,
  options?: RequestOptions
): Promise<OrderTrackingPageData | null> {
  try {
    const res = await shopifyFetch<ShopifyOrderTrackingPageOperation>({
      query: getOrderTrackingPageQuery,
      variables: { id },
      cache: 'no-store',
      headers: options?.authorization ? { Authorization: options.authorization } : {}
    });

    return {
      order: res.body.data.getOrder?.order ?? null,
      tracking: res.body.data.deliveryTracking ?? null
    };
  } catch (err) {
    console.error('[getOrderTrackingPage] Failed to load order tracking', { id, err });
    return null;
  }
}
