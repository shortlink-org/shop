'use client';

import { CourierStatus, STATUS_LABELS, STATUS_COLORS } from '@/types/courier';

interface CourierStatusBadgeProps {
  status: CourierStatus;
}

export function CourierStatusBadge({ status }: CourierStatusBadgeProps) {
  const toneClass =
    {
      default:
        'border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_70%,transparent)] text-[var(--color-muted-foreground)]',
      success:
        'border-[color-mix(in_srgb,var(--color-success)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-success)_12%,transparent)] text-[var(--color-success)]',
      processing:
        'border-[color-mix(in_srgb,var(--color-accent)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-accent)_12%,transparent)] text-[var(--color-accent)]',
      error:
        'border-[color-mix(in_srgb,var(--color-danger)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-danger)_12%,transparent)] text-[var(--color-danger)]',
    }[STATUS_COLORS[status]];

  return (
    <span
      className={`inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold tracking-wide ${toneClass}`}
    >
      {STATUS_LABELS[status]}
    </span>
  );
}
