'use client';

import clsx from 'clsx';
import { ShopSidebar } from 'components/layout/shop-sidebar';
import { ReactNode, useState } from 'react';

const shellBase =
  'mx-auto grid w-full max-w-screen-2xl gap-6 px-4 pb-8 pt-4 md:items-start md:gap-8 lg:px-6';

/**
 * Keeps the main column width in sync with ui-kit Sidebar full (~17rem) vs mini (7rem) mode.
 */
export function StorefrontLayoutShell({ children }: { children: ReactNode }) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <div
      className={clsx(
        shellBase,
        'md:transition-[grid-template-columns] md:duration-200 md:ease-out',
        sidebarCollapsed
          ? 'md:grid-cols-[minmax(0,7rem)_minmax(0,1fr)]'
          : 'md:grid-cols-[minmax(0,17rem)_minmax(0,1fr)]'
      )}
    >
      <aside className="hidden min-w-0 md:block">
        <ShopSidebar collapsed={sidebarCollapsed} onCollapsedChange={setSidebarCollapsed} />
      </aside>
      <div className="min-w-0">{children}</div>
    </div>
  );
}
