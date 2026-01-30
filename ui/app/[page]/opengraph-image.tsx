import OpengraphImage from 'components/opengraph-image';

export const runtime = 'edge';

export default async function Image(props: { params: Promise<{ page: string }> }) {
  const params = await props.params;
  const title = 'Page';

  return await OpengraphImage({ title });
}
