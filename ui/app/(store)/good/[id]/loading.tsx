export default function GoodPageLoading() {
  return (
    <div className="mx-auto max-w-screen-2xl px-4 pb-16">
      <div className="mb-4 flex flex-wrap items-center gap-2 sm:gap-4">
        <div className="h-8 w-24 animate-pulse rounded bg-[var(--color-muted)]" />
        <span className="hidden sm:inline">
          <div className="h-4 w-px bg-[var(--color-border)]" />
        </span>
        <div className="h-4 w-48 animate-pulse rounded bg-[var(--color-muted)]" />
      </div>
      <div className="grid gap-8 lg:grid-cols-2">
        <div className="aspect-square max-h-[550px] w-full animate-pulse rounded-xl bg-[var(--color-muted)]" />
        <div className="space-y-4">
          <div className="h-8 w-3/4 animate-pulse rounded bg-[var(--color-muted)]" />
          <div className="h-6 w-24 animate-pulse rounded bg-[var(--color-muted)]" />
          <div className="space-y-2 pt-4">
            <div className="h-3 w-full animate-pulse rounded bg-[var(--color-muted)]" />
            <div className="h-3 w-full animate-pulse rounded bg-[var(--color-muted)]" />
            <div className="h-3 w-2/3 animate-pulse rounded bg-[var(--color-muted)]" />
          </div>
          <div className="pt-6">
            <div className="h-12 w-40 animate-pulse rounded-full bg-[var(--color-muted)]" />
          </div>
        </div>
      </div>
    </div>
  );
}
