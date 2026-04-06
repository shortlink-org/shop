import '@/lib/temporal-polyfill';
import { CartProvider } from 'components/cart/cart-context';
import { Navbar } from 'components/layout/navbar';
import { Footer } from 'components/layout/footer';
import { Providers } from 'components/providers';
import { UiKitThemeBridge } from 'components/ui-kit-theme-bridge';
import { GeistSans } from 'geist/font/sans';
import { CART_UNAVAILABLE, getCart, type CartLoadResult } from 'lib/shopify';
import { ensureStartsWith, sanitizeJsonLd } from 'lib/utils';
import { headers } from 'next/headers';
import Script from 'next/script';
import { ReactNode, ViewTransition } from 'react';
import { Toaster } from 'sonner';
import { ThemeProvider } from 'next-themes';
import '@shortlink-org/ui-kit/dist/assets/index.css';
import './globals.css';

// DOCS: https://nextjs.org/docs/app/api-reference/file-conventions/route-segment-config#dynamic
export const dynamic = 'force-dynamic';

const { TWITTER_CREATOR, TWITTER_SITE, SITE_NAME } = process.env;
const baseUrl = process.env.NEXT_PUBLIC_VERCEL_URL
  ? `https://${process.env.NEXT_PUBLIC_VERCEL_URL}`
  : 'http://localhost:3000';
const twitterCreator = TWITTER_CREATOR ? ensureStartsWith(TWITTER_CREATOR, '@') : undefined;
const twitterSite = TWITTER_SITE ? ensureStartsWith(TWITTER_SITE, 'https://') : undefined;

export const metadata = {
  metadataBase: new URL(baseUrl),
  title: {
    default: SITE_NAME!,
    template: `%s | ${SITE_NAME}`
  },
  robots: {
    follow: true,
    index: true
  },
  ...(twitterCreator &&
    twitterSite && {
      twitter: {
        card: 'summary_large_image',
        creator: twitterCreator,
        site: twitterSite
      }
    })
};

export default async function RootLayout({ children }: { children: ReactNode }) {
  const requestHeaders = await headers();
  const authHeader = requestHeaders.get('authorization') ?? undefined;

  // Resolve the cart on the server so client cart state does not suspend or roll back between
  // an optimistic add-to-cart update and the revalidated snapshot.
  // Identity: only Authorization (JWT) is forwarded; oms-graphql gets x-user-id from Istio (RequestAuthentication outputClaimToHeaders).
  const initialCartResult: CartLoadResult = await getCart({
    authorization: authHeader
  }).catch(() => CART_UNAVAILABLE);

  const websiteJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: SITE_NAME,
    url: baseUrl
  };

  return (
    <html lang="en" className={GeistSans.variable} suppressHydrationWarning>
      <body
        className="bg-[var(--color-background)] text-[var(--color-foreground)] selection:bg-sky-200 selection:text-slate-950 dark:selection:bg-sky-500 dark:selection:text-slate-950"
        suppressHydrationWarning
      >
        <Script
          id="ld-json-website"
          type="application/ld+json"
          strategy="beforeInteractive"
          dangerouslySetInnerHTML={{ __html: sanitizeJsonLd(websiteJsonLd) }}
        />
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <UiKitThemeBridge />
          <Providers>
            <CartProvider initialCartResult={initialCartResult}>
              <Navbar />
              <main id="main" className="min-h-screen">
                <ViewTransition>{children}</ViewTransition>
                <Toaster closeButton />
              </main>
              <Footer />
            </CartProvider>
          </Providers>
        </ThemeProvider>
      </body>
    </html>
  );
}
