'use client';

import { useTheme } from 'next-themes';
import { useEffect } from 'react';

/**
 * @shortlink-org/ui-kit ToggleDarkMode (via AppHeader) dispatches `theme-toggle` CustomEvent;
 * next-themes does not listen by default — wire it to setTheme.
 */
export function UiKitThemeBridge() {
  const { setTheme } = useTheme();

  useEffect(() => {
    const onToggle = (ev: Event) => {
      const detail = (ev as CustomEvent<{ checked: boolean }>).detail;
      if (detail && typeof detail.checked === 'boolean') {
        setTheme(detail.checked ? 'dark' : 'light');
      }
    };
    window.addEventListener('theme-toggle', onToggle);
    return () => window.removeEventListener('theme-toggle', onToggle);
  }, [setTheme]);

  return null;
}
