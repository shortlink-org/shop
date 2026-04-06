import 'temporal-polyfill/global';
import { registerOTel } from '@vercel/otel';

/**
 * OpenTelemetry instrumentation for Next.js.
 * Runs once when the server starts. Enables automatic spans for routes, fetch, etc.
 * @see https://nextjs.org/docs/pages/guides/open-telemetry
 */
export function register() {
  registerOTel({
    serviceName: process.env.NEXT_PUBLIC_SERVICE_NAME ?? 'shortlink-shop-ui'
  });
}
