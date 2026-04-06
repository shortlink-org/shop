/**
 * pnpm patch for @shortlink-org/ui-kit failed integrity checks on this toolchain; apply tiny
 * storefront tweaks postinstall instead. Idempotent.
 */
const fs = require('fs');
const path = require('path');

const root = path.join(__dirname, '..');
const headerSearch = path.join(
  root,
  'node_modules/@shortlink-org/ui-kit/dist/page/AppHeader/components/HeaderSearch.js'
);

if (!fs.existsSync(headerSearch)) {
  process.exit(0);
}

let s = fs.readFileSync(headerSearch, 'utf8');
if (s.includes('min-w-0 flex-1 md:max-w-72 md:flex-none')) {
  process.exit(0);
}

const next = s
  .replace(
    'className: t("mr-1 hidden md:block", o),',
    'className: t("mr-1 min-w-0 flex-1 md:max-w-72 md:flex-none", o),'
  )
  .replace('className: "w-72",', 'className: "min-w-0 w-full",');

if (next === s) {
  process.exit(0);
}

fs.writeFileSync(headerSearch, next);
