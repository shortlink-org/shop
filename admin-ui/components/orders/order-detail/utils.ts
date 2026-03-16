import type { DeliveryAddress, DeliveryTracking, DeliveryTrackingLocation } from '@/types/order';

export function formatStatus(status?: string | null): string {
  if (!status) return 'Unknown';
  return status
    .toLowerCase()
    .split('_')
    .map((chunk) => chunk.charAt(0).toUpperCase() + chunk.slice(1))
    .join(' ');
}

export function formatDateTime(value?: string | null): string {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
}

export function formatMoney(value?: number | null): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return '—';
  return new Intl.NumberFormat('en', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 2
  }).format(value);
}

export function formatAddress(address?: DeliveryAddress | null): string {
  if (!address) return '—';
  return [address.street, address.city, address.country].filter(Boolean).join(', ') || '—';
}

export function formatDateTimeShort(value?: string | null): string {
  if (!value) return 'Just now';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Just now';
  return new Intl.DateTimeFormat('en', {
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
    day: 'numeric'
  }).format(date);
}

export function getStatusBadgeClass(status?: string | null): string {
  if (!status) {
    return 'border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_72%,transparent)] text-[var(--color-muted-foreground)]';
  }
  if (status === 'DELIVERED' || status === 'COMPLETED') {
    return 'border-[color-mix(in_srgb,var(--color-success)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-success)_12%,transparent)] text-[var(--color-success)]';
  }
  if (status === 'ASSIGNED' || status === 'IN_TRANSIT') {
    return 'border-[color-mix(in_srgb,var(--color-accent)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-accent)_12%,transparent)] text-[var(--color-accent)]';
  }
  if (status === 'NOT_DELIVERED' || status === 'REQUIRES_HANDLING' || status === 'CANCELLED') {
    return 'border-[color-mix(in_srgb,var(--color-danger)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-danger)_12%,transparent)] text-[var(--color-danger)]';
  }
  return 'border-[color-mix(in_srgb,var(--color-warning)_28%,transparent)] bg-[color-mix(in_srgb,var(--color-warning)_12%,transparent)] text-[var(--color-warning)]';
}

const TERMINAL_TRACKING_STATUSES = new Set(['DELIVERED', 'NOT_DELIVERED', 'REQUIRES_HANDLING']);

export function isTerminalTracking(status?: string | null): boolean {
  return Boolean(status && TERMINAL_TRACKING_STATUSES.has(status));
}

export function buildMiniMap(
  courierLocation?: DeliveryTrackingLocation | null,
  destination?: DeliveryAddress | null
) {
  const courierLat = courierLocation?.latitude;
  const courierLon = courierLocation?.longitude;
  const destinationLat = destination?.latitude;
  const destinationLon = destination?.longitude;

  if (
    typeof courierLat !== 'number' ||
    typeof courierLon !== 'number' ||
    typeof destinationLat !== 'number' ||
    typeof destinationLon !== 'number'
  ) {
    return null;
  }

  const latPadding = Math.max(Math.abs(courierLat - destinationLat) * 0.25, 0.01);
  const lonPadding = Math.max(Math.abs(courierLon - destinationLon) * 0.25, 0.01);
  const minLat = Math.min(courierLat, destinationLat) - latPadding;
  const maxLat = Math.max(courierLat, destinationLat) + latPadding;
  const minLon = Math.min(courierLon, destinationLon) - lonPadding;
  const maxLon = Math.max(courierLon, destinationLon) + lonPadding;

  const toPercentX = (lon: number) => ((lon - minLon) / (maxLon - minLon || 1)) * 100;
  const toPercentY = (lat: number) => (1 - (lat - minLat) / (maxLat - minLat || 1)) * 100;

  return {
    courier: {
      left: `${toPercentX(courierLon)}%`,
      top: `${toPercentY(courierLat)}%`
    },
    destination: {
      left: `${toPercentX(destinationLon)}%`,
      top: `${toPercentY(destinationLat)}%`
    }
  };
}
