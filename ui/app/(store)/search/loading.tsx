export default function SearchLoading() {
  return (
    <section className="shop-productgrid grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-6 lg:grid-cols-3 lg:gap-8">
      {Array.from({ length: 6 }).map((_, idx) => (
        <div
          key={idx}
          className="overflow-hidden rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)]"
        >
          <div className="aspect-[4/5] animate-pulse bg-[var(--color-muted)]" />
          <div className="space-y-2 p-4">
            <div className="h-4 w-3/4 animate-pulse rounded bg-[var(--color-muted)]" />
            <div className="h-3 w-full animate-pulse rounded bg-[var(--color-muted)]" />
            <div className="h-3 w-2/3 animate-pulse rounded bg-[var(--color-muted)]" />
          </div>
        </div>
      ))}
    </section>
  );
}
