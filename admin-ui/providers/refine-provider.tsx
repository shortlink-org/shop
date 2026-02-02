'use client';

import { Refine, DataProvider } from '@refinedev/core';
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
} from '@ant-design/icons';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

import { authProvider } from './auth-provider';
import { SessionWrapper } from '@/components/auth/SessionWrapper';
import { useSession } from '@/contexts/SessionContext';

import '@refinedev/antd/dist/reset.css';

const { Sider, Content, Header } = Layout;

interface RefineProviderProps {
  children: React.ReactNode;
}

// Mock data provider - replace with real GraphQL later
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const mockDataProvider: DataProvider = {
  getList: async ({ resource }): Promise<any> => {
    // Mock courier data
    if (resource === 'couriers') {
      return {
        data: [
          {
            id: '1',
            courierId: '1',
            name: 'Иван Петров',
            phone: '+79001234567',
            email: 'ivan@example.com',
            transportType: 'BICYCLE',
            status: 'FREE',
            rating: 4.8,
            workZone: 'Центр',
            currentLoad: 2,
            maxLoad: 5,
            successfulDeliveries: 156,
            failedDeliveries: 3,
          },
          {
            id: '2',
            courierId: '2',
            name: 'Мария Сидорова',
            phone: '+79007654321',
            email: 'maria@example.com',
            transportType: 'CAR',
            status: 'BUSY',
            rating: 4.9,
            workZone: 'Север',
            currentLoad: 4,
            maxLoad: 10,
            successfulDeliveries: 289,
            failedDeliveries: 5,
          },
          {
            id: '3',
            courierId: '3',
            name: 'Алексей Козлов',
            phone: '+79009876543',
            email: 'alex@example.com',
            transportType: 'MOTORCYCLE',
            status: 'UNAVAILABLE',
            rating: 4.5,
            workZone: 'Юг',
            currentLoad: 0,
            maxLoad: 8,
            successfulDeliveries: 98,
            failedDeliveries: 2,
          },
        ],
        total: 3,
      };
    }
    return { data: [], total: 0 };
  },
  getOne: async ({ resource, id }): Promise<any> => {
    if (resource === 'couriers') {
      return {
        data: {
          id,
          courierId: id,
          name: 'Иван Петров',
          phone: '+79001234567',
          email: 'ivan@example.com',
          transportType: 'BICYCLE',
          status: 'FREE',
          rating: 4.8,
          workZone: 'Центр',
          currentLoad: 2,
          maxLoad: 5,
          maxDistanceKm: 15,
          successfulDeliveries: 156,
          failedDeliveries: 3,
          workHours: {
            startTime: '09:00',
            endTime: '18:00',
            workDays: [1, 2, 3, 4, 5],
          },
          currentLocation: {
            latitude: 55.7558,
            longitude: 37.6173,
          },
          createdAt: '2024-01-15T10:00:00Z',
          lastActiveAt: '2024-02-02T14:30:00Z',
        },
      };
    }
    return { data: {} };
  },
  create: async ({ variables }): Promise<any> => {
    return { data: { id: 'new-id', ...variables } };
  },
  update: async ({ id, variables }): Promise<any> => {
    return { data: { id, ...variables } };
  },
  deleteOne: async ({ id }): Promise<any> => {
    return { data: { id } };
  },
  getApiUrl: () => process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:9991/graphql',
};

/**
 * User menu in header
 */
function UserMenu() {
  const { session, logout } = useSession();
  
  const userName = session?.identity?.traits 
    ? (session.identity.traits as Record<string, unknown>).name as string || 
      (session.identity.traits as Record<string, unknown>).email as string || 
      'Пользователь'
    : 'Пользователь';

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
        dataProvider={mockDataProvider}
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
  );
}
