import { Button } from '@/lib/ui-kit';
import { ClipboardDocumentListIcon, FireIcon, SparklesIcon } from '@heroicons/react/24/outline';
import { StorefrontHero } from 'components/layout/storefront-hero';
import { ShopLeaderboard } from 'components/layout/shop-leaderboard';
import { ShopProductGrid } from 'components/layout/shop-product-grid';
import { RetryButton } from 'components/retry-button';
import { getCollectionProducts, GOODS_UNAVAILABLE } from 'lib/shopify';
import { headers } from 'next/headers';
import Link from 'next/link';

export const metadata = {
  description: 'High-performance ecommerce store built with Next.js, Vercel, and Shopify.',
  openGraph: {
    type: 'website'
  }
};

export default async function HomePage(_props: {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  void _props; // Page signature; searchParams not used for homepage
  const authHeader = (await headers()).get('authorization') ?? undefined;
  // Never pass searchParams.page (or any URL param) into getCollectionProducts — BFF expects Int.
  // If we ever add pagination, parse page from searchParams and pass only a normalized integer.
  const homepageItems = await getCollectionProducts({ authorization: authHeader });

  if (homepageItems === GOODS_UNAVAILABLE) {
    return (
      <section className="w-full pb-4">
        <div className="flex flex-col items-center justify-center rounded-2xl border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_94%,transparent)] py-16 shadow-[0_16px_40px_-28px_rgba(15,23,42,0.18)] dark:shadow-[0_16px_48px_-28px_rgba(0,0,0,0.45)]">
          <p className="text-lg font-semibold text-[var(--color-foreground)]">We couldn&apos;t load products</p>
          <p className="mt-2 max-w-md text-center text-sm text-[var(--color-muted-foreground)]">
            We&apos;ll show them when they&apos;re available again.
          </p>
          <RetryButton />
        </div>
      </section>
    );
  }

  if (homepageItems.length === 0) return null;

  const spotlightCards = [
    {
      title: 'New arrivals',
      description: 'Fresh branded pieces and small-format drops for the next campaign wave.',
      href: '/search?sort=latest-desc',
      icon: SparklesIcon
    },
    {
      title: 'Best sellers',
      description: 'The fastest-moving goods, filtered for momentum and basket strength.',
      href: '/search?sort=trending-desc',
      icon: FireIcon
    },
    {
      title: 'Checkout flow',
      description: 'Move straight into basket review without losing the storefront visual context.',
      href: '/checkout',
      icon: ClipboardDocumentListIcon
    }
  ];

  return (
    <section className="mx-auto w-full max-w-7xl pt-2 pb-14 sm:pt-4">
      <div className="space-y-8">
        <StorefrontHero goods={homepageItems.slice(0, 6)} />

        <div className="space-y-8">
          <section className="shop-panel rounded-[2rem] border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_92%,transparent)] p-4 shadow-[0_32px_90px_-54px_rgba(15,23,42,0.38)] sm:p-6">
              <div className="flex flex-col gap-4 border-b border-[var(--color-border)] pb-5 sm:flex-row sm:items-end sm:justify-between">
                <div className="max-w-2xl">
                  <p className="text-[11px] font-semibold tracking-[0.2em] text-[var(--color-muted-foreground)] uppercase">
                    Featured shelf
                  </p>
                  <h2 className="mt-2 text-2xl font-semibold tracking-tight text-[var(--color-foreground)] sm:text-[2rem]">
                    A tighter, more editorial storefront grid
                  </h2>
                  <p className="mt-3 text-sm leading-7 text-[var(--color-muted-foreground)] sm:text-base">
                    Product cards now sit inside a cleaner shell, with consistent artwork and a
                    stronger hierarchy between product, price and quick actions.
                  </p>
                </div>
                <Button as={Link} asProps={{ href: '/search' }} variant="secondary">
                  Browse full catalog
                </Button>
              </div>

              <div className="pt-2">
                <ShopProductGrid
                  goods={homepageItems}
                  className="shop-home-grid"
                  gridClassName="gap-5 lg:gap-6"
                />
              </div>
            </section>

            <section className="grid gap-4 md:grid-cols-3">
              {spotlightCards.map((card) => {
                const Icon = card.icon;

                return (
                  <article
                    key={card.title}
                    className="rounded-[1.6rem] border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_88%,transparent)] p-5 shadow-[0_24px_60px_-46px_rgba(15,23,42,0.34)]"
                  >
                    <div className="flex size-12 items-center justify-center rounded-2xl bg-[var(--color-muted)] text-[var(--color-foreground)]">
                      <Icon className="size-5" />
                    </div>
                    <h3 className="mt-4 text-lg font-semibold tracking-tight text-[var(--color-foreground)]">
                      {card.title}
                    </h3>
                    <p className="mt-3 text-sm leading-7 text-[var(--color-muted-foreground)]">
                      {card.description}
                    </p>
                    <div className="mt-5">
                      <Button as={Link} asProps={{ href: card.href }} variant="outline" size="sm">
                        Open section
                      </Button>
                    </div>
                  </article>
                );
              })}
            </section>

            <div id="leaderboard" className="shop-panel">
              <ShopLeaderboard />
            </div>
        </div>
      </div>
    </section>
  );
}
