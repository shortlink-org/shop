import { Button, StatCard } from '@/lib/ui-kit';
import { ArrowTrendingUpIcon, SparklesIcon, SwatchIcon } from '@heroicons/react/24/outline';
import { Good } from 'lib/shopify/types';
import { getStorefrontArtwork, getStorefrontCategory } from 'lib/storefront-art';
import Link from 'next/link';

function formatPrice(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0
  }).format(amount);
}

function averagePrice(goods: Good[]): number {
  if (goods.length === 0) return 0;

  return goods.reduce((total, good) => total + good.price, 0) / goods.length;
}

function truncateDescription(description: string, maxLength = 120): string {
  if (description.length <= maxLength) return description;
  return `${description.slice(0, maxLength).trim()}...`;
}

export function StorefrontHero({ goods }: { goods: Good[] }) {
  const featuredGood = goods[0];
  const averageTicket = averagePrice(goods);
  const highestPrice = goods.reduce((highest, good) => Math.max(highest, good.price), 0);

  if (!featuredGood) {
    return null;
  }

  const artwork = getStorefrontArtwork(featuredGood.name, featuredGood.id, {
    width: 520,
    height: 400,
    eyebrow: 'featured drop',
    subtitle: getStorefrontCategory(featuredGood.name)
  });

  return (
    <section className="shop-hero relative overflow-hidden rounded-[2rem] border border-[var(--color-border)] bg-[linear-gradient(135deg,rgba(255,255,255,0.92),rgba(226,232,240,0.72))] p-4 shadow-[0_40px_100px_-58px_rgba(15,23,42,0.45)] backdrop-blur-xl sm:p-5 lg:p-6 dark:bg-[linear-gradient(135deg,rgba(15,23,42,0.92),rgba(15,23,42,0.72))]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(56,189,248,0.16),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(251,191,36,0.12),transparent_24%)]" />
      <div className="relative grid gap-4 lg:grid-cols-[minmax(0,1.1fr)_minmax(280px,0.72fr)] lg:items-start lg:gap-5">
        <div className="flex flex-col gap-4 sm:gap-5">
          <div className="max-w-2xl">
            <div className="inline-flex items-center gap-2 rounded-full border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_82%,transparent)] px-2.5 py-1 text-[10px] font-semibold tracking-[0.2em] text-[var(--color-muted-foreground)] uppercase">
              <SparklesIcon className="size-3.5 text-sky-500" />
              Curated storefront
            </div>
            <h1 className="mt-3 max-w-2xl text-2xl font-semibold tracking-[-0.03em] text-[var(--color-foreground)] sm:text-3xl lg:text-4xl">
              Merch that reads like a campaign, not a commodity shelf.
            </h1>
            <p className="mt-2 max-w-xl text-sm leading-6 text-[var(--color-muted-foreground)] sm:text-base sm:leading-7">
              The shop now leans into the visual language from `ui-kit`: editorial surfaces,
              stronger hierarchy and a cleaner path from discovery to checkout.
            </p>
            <div className="mt-4 flex flex-wrap items-center gap-2">
              <Button as={Link} asProps={{ href: '/search' }}>
                Explore catalog
              </Button>
              <Button as={Link} asProps={{ href: '/#leaderboard' }} variant="secondary">
                See leaderboard
              </Button>
            </div>
          </div>

          <div className="grid gap-2 sm:grid-cols-3">
            <StatCard
              label="Featured goods"
              value={goods.length.toString().padStart(2, '0')}
              change="live assortment"
              tone="accent"
            />
            <StatCard
              label="Average ticket"
              value={formatPrice(averageTicket)}
              change="curated mix"
              tone="success"
            />
            <StatCard
              label="Hero price"
              value={formatPrice(highestPrice)}
              change="premium shelf"
              tone="warning"
            />
          </div>
        </div>

        <div className="grid gap-3">
          <div className="rounded-[1.5rem] border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_90%,transparent)] p-3 shadow-[0_28px_70px_-52px_rgba(15,23,42,0.45)]">
            {/* eslint-disable-next-line @next/next/no-img-element -- dynamic artwork URL from getStorefrontArtwork */}
            <img
              src={artwork}
              alt={`${featuredGood.name} artwork`}
              className="aspect-[4/3] max-h-[200px] w-full rounded-[1.15rem] border border-white/10 object-cover shadow-[0_24px_64px_-40px_rgba(15,23,42,0.62)] sm:max-h-[220px]"
            />
            <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="min-w-0">
                <p className="text-[10px] font-semibold tracking-[0.18em] text-[var(--color-muted-foreground)] uppercase">
                  Featured drop
                </p>
                <h2 className="mt-1 text-lg font-semibold tracking-tight text-[var(--color-foreground)]">
                  {featuredGood.name}
                </h2>
                <p className="mt-1 line-clamp-2 text-xs leading-5 text-[var(--color-muted-foreground)]">
                  {truncateDescription(featuredGood.description, 90)}
                </p>
              </div>
              <Button
                as={Link}
                asProps={{ href: `/good/${featuredGood.id}` }}
                variant="outline"
                className="shrink-0"
              >
                View item
              </Button>
            </div>
          </div>

          <div className="grid gap-2 sm:grid-cols-2">
            <div className="rounded-xl border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_88%,transparent)] p-3">
              <div className="flex items-center gap-2">
                <div className="flex size-9 items-center justify-center rounded-xl bg-sky-100 text-sky-700 dark:bg-sky-950/40 dark:text-sky-200">
                  <ArrowTrendingUpIcon className="size-4" />
                </div>
                <div>
                  <p className="text-[10px] font-semibold tracking-[0.14em] text-[var(--color-muted-foreground)] uppercase">
                    Momentum
                  </p>
                  <p className="text-xs font-semibold text-[var(--color-foreground)]">
                    Trending shelf
                  </p>
                </div>
              </div>
              <p className="mt-2 line-clamp-2 text-xs leading-5 text-[var(--color-muted-foreground)]">
                Showcase the fastest-moving items first and keep the leaderboard close to the buy
                flow.
              </p>
            </div>
            <div className="rounded-xl border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_88%,transparent)] p-3">
              <div className="flex items-center gap-2">
                <div className="flex size-9 items-center justify-center rounded-xl bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-200">
                  <SwatchIcon className="size-4" />
                </div>
                <div>
                  <p className="text-[10px] font-semibold tracking-[0.14em] text-[var(--color-muted-foreground)] uppercase">
                    Brand system
                  </p>
                  <p className="text-xs font-semibold text-[var(--color-foreground)]">
                    Cohesive product art
                  </p>
                </div>
              </div>
              <p className="mt-2 line-clamp-2 text-xs leading-5 text-[var(--color-muted-foreground)]">
                Consistent poster-style visuals so the catalog feels designed even without uploaded
                media.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
