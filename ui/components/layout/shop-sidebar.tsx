'use client';

import { Sidebar } from '@/lib/ui-kit';
import { ChevronDoubleLeftIcon, ChevronDoubleRightIcon } from '@heroicons/react/24/outline';
import { shopSidebarSections } from 'lib/shop-navigation';
import { usePathname } from 'next/navigation';

export type ShopSidebarProps = {
  collapsed: boolean;
  onCollapsedChange: (collapsed: boolean) => void;
};

export function ShopSidebar({ collapsed, onCollapsedChange }: ShopSidebarProps) {
  const pathname = usePathname();

  return (
    <Sidebar
      sections={shopSidebarSections()}
      activePath={pathname}
      collapsed={collapsed}
      onCollapsedChange={onCollapsedChange}
      density="compact"
      variant="sticky"
      height="calc(100vh - 6.5rem)"
      className="shop-sidebar top-8"
      footerSlot={
        <div className="border-t border-[var(--color-border)]/80 px-2.5 pb-3 pt-2">
          <div className="flex justify-center">
            <button
              type="button"
              onClick={() => onCollapsedChange(!collapsed)}
              aria-expanded={!collapsed}
              aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
              title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
              className="focus-ring inline-flex size-10 shrink-0 cursor-pointer items-center justify-center rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] text-[var(--color-muted-foreground)] transition-colors duration-200 hover:bg-[var(--color-muted)] hover:text-[var(--color-foreground)]"
            >
              {collapsed ? (
                <ChevronDoubleRightIcon className="size-4" aria-hidden />
              ) : (
                <ChevronDoubleLeftIcon className="size-4" aria-hidden />
              )}
            </button>
          </div>
        </div>
      }
    />
  );
}
