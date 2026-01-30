import OpengraphImage from 'components/opengraph-image';
import { getCollection } from 'lib/shopify';

export const runtime = 'edge';

export default async function Image(props: { params: Promise<{ collection: string }> }) {
  const params = await props.params;
  const collection = await getCollection(Number(params.collection));
  const title = collection?.seo?.title || collection?.title;

  return await OpengraphImage({ title });
}
