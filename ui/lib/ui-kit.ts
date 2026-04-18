// Load polyfill before the ui-kit barrel (Footer evaluates Temporal at module scope).
import '@/lib/temporal-polyfill';
export * from '@shortlink-org/ui-kit';
