/**
 * Published `@shortlink-org/ui-kit` declares `"types": "dist/index.d.ts"` but the tarball ships no `.d.ts` files.
 * This satisfies `export *` in `lib/ui-kit.ts`; public types for `@/lib/ui-kit` remain in `types/ui-kit.d.ts`.
 */
declare module '@shortlink-org/ui-kit';
