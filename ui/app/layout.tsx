import { CartProvider } from 'components/cart/cart-context';
import { Navbar } from 'components/layout/navbar';
import { Footer } from 'components/layout/footer';
import { Providers } from 'components/providers';
import { GeistSans } from 'geist/font/sans';
import { CART_UNAVAILABLE, getCart, type CartLoadResult } from 'lib/shopify';
import { ensureStartsWith, sanitizeJsonLd } from 'lib/utils';
import { headers } from 'next/headers';
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

  // Pass cart promise that never rejects — when carts service is down we resolve with CART_UNAVAILABLE so UI can show "we'll display it later"
  // Identity: only Authorization (JWT) is forwarded; oms-graphql gets x-user-id from Istio (RequestAuthentication outputClaimToHeaders).
  const cartPromise: Promise<CartLoadResult> = getCart({
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
      <body className="bg-[var(--color-background)] text-[var(--color-foreground)] selection:bg-sky-200 selection:text-slate-950 dark:selection:bg-sky-500 dark:selection:text-slate-950">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: sanitizeJsonLd(websiteJsonLd) }}
        />
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <Providers>
            <CartProvider cartPromise={cartPromise}>
              <a
                href="#main"
                className="focus-ring absolute top-4 left-4 z-[100] -translate-y-full rounded-md bg-[var(--color-foreground)] px-4 py-2 text-sm font-medium text-[var(--color-background)] transition-transform focus:translate-y-0"
              >
                Skip to main content
              </a>
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
