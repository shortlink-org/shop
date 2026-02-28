'use client';

import { ApolloProvider } from '@apollo/client/react';
import { Refine } from '@refinedev/core';
import { RefineKbar, RefineKbarProvider } from '@refinedev/kbar';
import { useNotificationProvider } from '@refinedev/antd';
import routerProvider from '@refinedev/nextjs-router';
import { App as AntdApp, ConfigProvider, Layout, Menu, theme, Avatar, Dropdown, Space } from 'antd';
import type { MenuProps } from 'antd';
import ruRU from 'antd/locale/ru_RU';
import { 
  TeamOutlined, 
  DashboardOutlined, 
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  SwapOutlined,
} from '@ant-design/icons';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

import { authProvider } from './auth-provider';
import { graphqlDataProvider } from '@/lib/graphql-data-provider';
import { apolloClient } from '@/lib/apollo-client';
import { SessionWrapper } from '@/components/auth/SessionWrapper';
import { useSession } from '@/contexts/SessionContext';
import { getUserName } from '@/lib/ory/api';

import '@refinedev/antd/dist/reset.css';

const { Sider, Content, Header } = Layout;

interface RefineProviderProps {
  children: React.ReactNode;
}

/**
 * User menu in header
 */
function UserMenu() {
  const { session, logout } = useSession();
  
  const userName = session ? getUserName(session) : 'Пользователь';

  const items: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Профиль',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: 'Настройки',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Выйти',
      danger: true,
      onClick: () => logout(),
    },
  ];

  return (
    <Dropdown menu={{ items }} placement="bottomRight">
      <Space className="cursor-pointer">
        <Avatar icon={<UserOutlined />} />
        <span className="hidden md:inline">{userName}</span>
      </Space>
    </Dropdown>
  );
}

function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  
  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: <Link href="/">Дашборд</Link>,
    },
    {
      key: '/couriers',
      icon: <TeamOutlined />,
      label: <Link href="/couriers">Курьеры</Link>,
    },
    {
      key: 'django-admin',
      icon: <SwapOutlined />,
      label: <a href="/admin">Django Admin</a>,
    },
  ];

  const selectedKey = pathname?.startsWith('/couriers') ? '/couriers' : '/';

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        breakpoint="lg"
        collapsedWidth="0"
        style={{ background: '#001529' }}
      >
        <div className="p-4 text-white text-xl font-bold text-center">
          Delivery Admin
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Header style={{ 
          padding: '0 24px', 
          background: '#fff',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <span className="text-lg font-semibold">Панель управления доставкой</span>
          <UserMenu />
        </Header>
        <Content style={{ margin: '0', background: '#f5f5f5' }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
}

function RefineApp({ children }: { children: React.ReactNode }) {
  return (
    <RefineKbarProvider>
      <Refine
        routerProvider={routerProvider}
        dataProvider={graphqlDataProvider}
        authProvider={authProvider}
        notificationProvider={useNotificationProvider}
        resources={[
          {
            name: 'dashboard',
            list: '/',
            meta: {
              label: 'Дашборд',
              icon: <DashboardOutlined />,
            },
          },
          {
            name: 'couriers',
            list: '/couriers',
            show: '/couriers/:id',
            create: '/couriers/create',
            edit: '/couriers/:id/edit',
            meta: {
              label: 'Курьеры',
              icon: <TeamOutlined />,
            },
          },
        ]}
        options={{
          syncWithLocation: true,
          warnWhenUnsavedChanges: true,
        }}
      >
        <AdminLayout>{children}</AdminLayout>
        <RefineKbar />
      </Refine>
    </RefineKbarProvider>
  );
}

export function RefineProvider({ children }: RefineProviderProps) {
  return (
    <ApolloProvider client={apolloClient}>
      <ConfigProvider
        locale={ruRU}
        theme={{
          algorithm: theme.defaultAlgorithm,
        }}
      >
        <AntdApp>
          <SessionWrapper requireAuth={false}>
            <RefineApp>{children}</RefineApp>
          </SessionWrapper>
        </AntdApp>
      </ConfigProvider>
    </ApolloProvider>
  );
}
