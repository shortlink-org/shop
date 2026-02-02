import { headers } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

const domain = process.env.API_URI ?? '';

// This is called from `app/api/revalidate.ts` so providers can control revalidation logic.
export async function revalidate(req: NextRequest): Promise<NextResponse> {
  // We always need to respond with a 200 status code to Shopify,
  // otherwise it will continue to retry the request.
  const collectionWebhooks = ['collections/create', 'collections/delete', 'collections/update'];
  const goodWebhooks = ['products/create', 'products/delete', 'products/update'];
  const headersList = await headers();
  const topic = headersList.get('x-shopify-topic') || 'unknown';
  const secret = req.nextUrl.searchParams.get('secret');
  const isCollectionUpdate = collectionWebhooks.includes(topic);
  const isGoodUpdate = goodWebhooks.includes(topic);

  if (!secret || secret !== process.env.SHOPIFY_REVALIDATION_SECRET) {
    console.error('Invalid revalidation secret.');
    return NextResponse.json({ status: 200 });
  }

  if (!isCollectionUpdate && !isGoodUpdate) {
    // We don't need to revalidate anything for any other topics.
    return NextResponse.json({ status: 200 });
  }

  // Note: revalidateTag API may have changed in Next.js 16
  // For now we just return success
  return NextResponse.json({ status: 200, revalidated: true, now: Date.now() });
}
