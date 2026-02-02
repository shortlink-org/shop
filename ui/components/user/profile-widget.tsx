'use client';

import { Menu, Transition } from '@headlessui/react';
import { AxiosError } from 'axios';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState, useEffect, Fragment } from 'react';
import { UserIcon, ArrowRightOnRectangleIcon, ShoppingBagIcon } from '@heroicons/react/24/outline';

import { useSession } from '@/contexts/SessionContext';
import ory from '@/lib/ory/sdk';
import clsx from 'clsx';

export default function ProfileWidget() {
  const [logoutToken, setLogoutToken] = useState<string>('');
  const router = useRouter();
  const { session, hasSession, isLoading } = useSession();

  const traits: Record<string, unknown> = (session?.identity?.traits as Record<string, unknown>) ?? {};
  const nameTraits = traits?.name as Record<string, string> | undefined;
  const firstName = nameTraits?.first ?? '';
  const lastName = nameTraits?.last ?? '';
  const email = (traits?.email as string) ?? '';
  const displayName = `${firstName} ${lastName}`.trim() || email || 'User';
  const initials =
    (firstName?.[0] ?? '') + (lastName?.[0] ?? '') || (email?.[0] ?? '').toUpperCase() || 'U';

  useEffect(() => {
    if (!hasSession) return;

    ory
      .createBrowserLogoutFlow()
      .then(({ data }) => {
        setLogoutToken(data.logout_token);
      })
      .catch((err: AxiosError) => {
        if (err.response?.status === 401) return;
        return Promise.reject(err);
      });
  }, [hasSession]);

  if (isLoading) {
    return <div className="h-8 w-8 animate-pulse rounded-full bg-neutral-200 dark:bg-neutral-700" />;
  }

  if (!hasSession) {
    return (
      <Link
        href="/auth/login"
        className="rounded-md px-4 py-2 text-sm font-medium text-neutral-900 transition-opacity hover:opacity-80 dark:text-white"
      >
        Sign in
      </Link>
    );
  }

  const menuItems = [
    {
      name: 'Your Profile',
      link: '/profile',
      icon: UserIcon,
    },
    {
      name: 'Orders',
      link: '/orders',
      icon: ShoppingBagIcon,
    },
    {
      name: 'Sign out',
      link: '#',
      icon: ArrowRightOnRectangleIcon,
      onClick: () => {
        ory
          .updateLogoutFlow({ token: logoutToken })
          .then(() => router.push('/auth/login'))
          .then(() => window.location.reload());
      },
    },
  ];

  return (
    <Menu as="div" className="relative">
      {({ open }) => (
        <>
          <Menu.Button className="flex items-center gap-2 rounded-full p-1.5 transition-all hover:bg-neutral-100 dark:hover:bg-neutral-800">
            <span className="sr-only">Open user menu</span>
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-teal-500 text-xs font-semibold text-white">
              {initials}
            </div>
          </Menu.Button>

          <Transition
            show={open}
            as={Fragment}
            enter="transition ease-out duration-200"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-150"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className="absolute right-0 z-50 mt-2 w-56 origin-top-right rounded-xl bg-white shadow-lg ring-1 ring-black/5 focus:outline-none dark:bg-neutral-800 dark:ring-white/10">
              <div className="border-b border-neutral-100 px-4 py-3 dark:border-neutral-700">
                <div className="truncate text-sm font-semibold text-neutral-900 dark:text-white">
                  {displayName}
                </div>
                {email && (
                  <div className="truncate text-xs text-neutral-500 dark:text-neutral-400">
                    {email}
                  </div>
                )}
              </div>
              <div className="p-1">
                {menuItems.map((item) => (
                  <Menu.Item key={item.name}>
                    {({ active }) => {
                      const Icon = item.icon;
                      const content = (
                        <div
                          className={clsx(
                            active ? 'bg-neutral-50 dark:bg-neutral-700' : '',
                            'flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2'
                          )}
                          onClick={item.onClick}
                        >
                          <Icon className="h-5 w-5 text-neutral-500 dark:text-neutral-400" />
                          <span className="text-sm text-neutral-700 dark:text-neutral-200">
                            {item.name}
                          </span>
                        </div>
                      );

                      if (item.onClick) {
                        return content;
                      }

                      return (
                        <Link href={item.link} className="block">
                          {content}
                        </Link>
                      );
                    }}
                  </Menu.Item>
                ))}
              </div>
            </Menu.Items>
          </Transition>
        </>
      )}
    </Menu>
  );
}
