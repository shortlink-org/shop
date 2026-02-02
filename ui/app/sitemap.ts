import { getCollections, getPages, getGoods, GOODS_UNAVAILABLE } from 'lib/shopify';
import { validateEnvironmentVariables } from 'lib/utils';
import { MetadataRoute } from 'next';

type Route = {
  url: string;
  lastModified: string;
};

const baseUrl = process.env.NEXT_PUBLIC_VERCEL_URL
  ? `https://${process.env.NEXT_PUBLIC_VERCEL_URL}`
  : 'http://localhost:3000';

export const dynamic = 'force-dynamic';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  validateEnvironmentVariables();

  const routesMap = [''].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: new Date().toISOString()
  }));

  const collectionsPromise = getCollections().then((collections) =>
    collections === GOODS_UNAVAILABLE
      ? []
      : collections.map((collection) => ({
          url: `${baseUrl}${collection.path}`,
          lastModified: collection.updatedAt
        }))
  );

  const productsPromise = getGoods({}).then((products) =>
    products === GOODS_UNAVAILABLE
      ? []
      : products.map((product) => ({
          url: `${baseUrl}/good/${product.id}`,
          lastModified: product.updatedAt
        }))
  );

  const pagesPromise = getPages().then((pages) =>
    pages.map((page) => ({
      url: `${baseUrl}/page/${page.id}`,
      lastModified: page.updatedAt
    }))
  );

  let fetchedRoutes: Route[] = [];

  try {
    fetchedRoutes = (await Promise.all([collectionsPromise, productsPromise, pagesPromise])).flat();
  } catch (error) {
    throw JSON.stringify(error, null, 2);
  }

  return [...routesMap, ...fetchedRoutes];
}
