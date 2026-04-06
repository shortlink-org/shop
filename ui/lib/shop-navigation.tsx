import type { SidebarSection } from '@/lib/ui-kit';
import {
  ClipboardDocumentListIcon,
  HomeIcon,
  ShoppingBagIcon
} from '@heroicons/react/24/outline';
import type { ComponentType, SVGProps } from 'react';

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>;

export type ShopNavItem = {
  name: string;
  href: string;
  icon: IconComponent;
};

/** Primary storefront links — same entries in header rail and sidebar. */
export const SHOP_PRIMARY_NAV: ShopNavItem[] = [
  { name: 'Home', href: '/', icon: HomeIcon },
  { name: 'Shop', href: '/search', icon: ShoppingBagIcon },
  { name: 'Checkout', href: '/checkout', icon: ClipboardDocumentListIcon }
];

export function shopHeaderNavigation(): { name: string; href: string }[] {
  return SHOP_PRIMARY_NAV.map(({ name, href }) => ({ name, href }));
}

export function shopSidebarSections(): SidebarSection[] {
  return [
    {
      type: 'simple',
      items: SHOP_PRIMARY_NAV.map(({ name, href, icon: Icon }) => ({
        name,
        url: href,
        icon: <Icon className="size-5 shrink-0" aria-hidden />
      }))
    }
  ];
}
