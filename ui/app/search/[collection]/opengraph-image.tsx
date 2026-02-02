import OpengraphImage from 'components/opengraph-image';
import { getCollection, GOODS_UNAVAILABLE } from 'lib/shopify';

export const runtime = 'edge';

export default async function Image(props: { params: Promise<{ collection: string }> }) {
  const params = await props.params;
  const collection = await getCollection(Number(params.collection));
  const title =
    collection === GOODS_UNAVAILABLE
      ? 'Collection'
      : collection?.seo?.title || collection?.title;

  return await OpengraphImage({ title });
}
