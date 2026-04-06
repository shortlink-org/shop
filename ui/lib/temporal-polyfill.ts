/**
 * Installs `Temporal` on globalThis for code that expects a global (e.g. @shortlink-org/ui-kit Footer).
 * Lighter than @js-temporal/polyfill and supports ESM global via the `/global` entry.
 * @see https://github.com/fullcalendar/temporal-polyfill
 */
import 'temporal-polyfill/global';
