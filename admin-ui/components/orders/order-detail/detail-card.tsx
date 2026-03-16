function DetailCard({
  title,
  children
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="admin-card p-6">
      <h2 className="text-lg font-semibold tracking-tight">{title}</h2>
      <div className="mt-5 space-y-4">{children}</div>
    </section>
  );
}

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="grid gap-2 border-b border-[var(--color-border)] pb-4 last:border-b-0 last:pb-0 sm:grid-cols-[180px_minmax(0,1fr)]">
      <p className="text-sm font-medium text-[var(--color-muted-foreground)]">{label}</p>
      <div className="text-sm text-[var(--color-foreground)]">{value}</div>
    </div>
  );
}

export { DetailCard, DetailRow };
