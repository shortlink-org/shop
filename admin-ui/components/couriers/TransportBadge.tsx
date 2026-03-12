'use client';

import { TransportType, TRANSPORT_LABELS } from '@/types/courier';

interface TransportBadgeProps {
  type: TransportType;
}

const TRANSPORT_ICONS: Record<TransportType, string> = {
  UNSPECIFIED: '•',
  WALKING: 'W',
  BICYCLE: 'B',
  MOTORCYCLE: 'M',
  CAR: 'C',
};

const TRANSPORT_CLASSES: Record<TransportType, string> = {
  UNSPECIFIED:
    'border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_70%,transparent)] text-[var(--color-muted-foreground)]',
  WALKING:
    'border-[color-mix(in_srgb,var(--color-success)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-success)_12%,transparent)] text-[var(--color-success)]',
  BICYCLE:
    'border-[color-mix(in_srgb,var(--color-accent)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-accent)_12%,transparent)] text-[var(--color-accent)]',
  MOTORCYCLE:
    'border-[color-mix(in_srgb,var(--color-warning)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-warning)_12%,transparent)] text-[var(--color-warning)]',
  CAR: 'border-[color-mix(in_srgb,#6366f1_28%,transparent)] bg-[color-mix(in_srgb,#6366f1_12%,transparent)] text-[#6366f1]',
};

export function TransportBadge({ type }: TransportBadgeProps) {
  return (
    <span
      className={`inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-xs font-semibold tracking-wide ${TRANSPORT_CLASSES[type]}`}
    >
      <span className="inline-flex h-4 w-4 items-center justify-center rounded-full border border-current/20 text-[10px]">
        {TRANSPORT_ICONS[type]}
      </span>
      {TRANSPORT_LABELS[type]}
    </span>
  );
}
