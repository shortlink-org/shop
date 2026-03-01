import { CartProvider } from 'components/cart/cart-context';
import { Navbar } from 'components/layout/navbar';
import { Footer } from 'components/layout/footer';
import { Providers } from 'components/providers';
import { GeistSans } from 'geist/font/sans';
import { CART_UNAVAILABLE, getCart, type CartLoadResult } from 'lib/shopify';
import { ensureStartsWith } from 'lib/utils';
import { cookies, headers } from 'next/headers';
import { ReactNode, ViewTransition } from 'react';
import { Toaster } from 'sonner';
import { ThemeProvider } from 'next-themes';
import './globals.css';

// DOCS: https://nextjs.org/docs/app/api-reference/file-conventions/route-segment-config#dynamic
export const dynamic = 'force-dynamic'

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
  const cookieStore = await cookies();
  const cartId = cookieStore.get('cartId')?.value;
  const authHeader = (await headers()).get('authorization') ?? undefined;

  // Pass cart promise that never rejects — when carts service is down we resolve with CART_UNAVAILABLE so UI can show "we'll display it later"
  const cartPromise: Promise<CartLoadResult> = getCart(cartId, {
    authorization: authHeader
  }).catch(() => CART_UNAVAILABLE);

  return (
    <html lang="en" className={GeistSans.variable} suppressHydrationWarning>
      <body className="bg-neutral-50 text-black selection:bg-teal-300 dark:bg-neutral-900 dark:text-white dark:selection:bg-pink-500 dark:selection:text-white">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <Providers>
            <CartProvider cartPromise={cartPromise}>
              <Navbar />
              <main className="min-h-screen">
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
