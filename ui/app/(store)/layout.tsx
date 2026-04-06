import { StorefrontLayoutShell } from 'components/layout/storefront-layout-shell';
import { ReactNode } from 'react';

/**
 * Shared storefront chrome: primary navigation sidebar (md+) matches the header menu.
 * Checkout and other flows stay outside this group.
 */
export default function StoreLayout({ children }: { children: ReactNode }) {
  return <StorefrontLayoutShell>{children}</StorefrontLayoutShell>;
}
