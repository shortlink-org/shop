// Load polyfill before the ui-kit barrel (Footer evaluates Temporal at module scope).
import '@/lib/temporal-polyfill';
export * from '../node_modules/@shortlink-org/ui-kit/dist/index.js';
