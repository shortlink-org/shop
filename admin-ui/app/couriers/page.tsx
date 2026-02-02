'use client';

import { useState } from 'react';
import { useList } from '@refinedev/core';
import { 
  Table, 
  Card, 
  Button, 
  Space, 
  Select, 
  Rate,
  Tooltip,
  message,
  Popconfirm,
} from 'antd';
import { 
  PlusOutlined, 
  ReloadOutlined,
  CheckOutlined,
  StopOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import Link from 'next/link';
import type { ColumnsType } from 'antd/es/table';

import { CourierStatusBadge } from '@/components/couriers/CourierStatusBadge';
import { TransportBadge } from '@/components/couriers/TransportBadge';
import type { Courier, CourierStatus, TransportType } from '@/types/courier';

export default function CouriersListPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState<CourierStatus[]>([]);
  const [transportFilter, setTransportFilter] = useState<TransportType[]>([]);

  const { query, result } = useList<Courier>({
    resource: 'couriers',
  });
  
  const { isLoading, refetch } = query;
  const data = result;

  const handleActivate = (id: string) => {
    message.success('Курьер активирован (mock)');
    refetch();
  };

  const handleDeactivate = (id: string) => {
    message.success('Курьер деактивирован (mock)');
    refetch();
  };

  const columns: ColumnsType<Courier> = [
    {
      title: 'Имя',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Courier) => (
        <Link href={`/couriers/${record.courierId}`}>
          <span className="text-blue-600 hover:underline">{name}</span>
        </Link>
      ),
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: CourierStatus) => <CourierStatusBadge status={status} />,
    },
    {
      title: 'Транспорт',
      dataIndex: 'transportType',
      key: 'transportType',
      render: (type: TransportType) => <TransportBadge type={type} />,
    },
    {
      title: 'Зона',
      dataIndex: 'workZone',
      key: 'workZone',
    },
    {
      title: 'Рейтинг',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating: number) => (
        <Rate disabled defaultValue={rating} allowHalf />
      ),
    },
    {
      title: 'Загрузка',
      key: 'load',
      render: (_: unknown, record: Courier) => (
        <span>
          {record.currentLoad} / {record.maxLoad}
        </span>
      ),
    },
    {
      title: 'Доставки',
      key: 'deliveries',
      render: (_: unknown, record: Courier) => (
        <Tooltip title={`Успешных: ${record.successfulDeliveries}, Неудачных: ${record.failedDeliveries}`}>
          <span className="text-green-600">{record.successfulDeliveries}</span>
          {' / '}
          <span className="text-red-600">{record.failedDeliveries}</span>
        </Tooltip>
      ),
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_: unknown, record: Courier) => (
        <Space size="small">
          <Tooltip title="Просмотр">
            <Link href={`/couriers/${record.courierId}`}>
              <Button icon={<EyeOutlined />} size="small" />
            </Link>
          </Tooltip>
          
          {record.status === 'UNAVAILABLE' && (
            <Tooltip title="Активировать">
              <Popconfirm
                title="Активировать курьера?"
                onConfirm={() => handleActivate(record.courierId)}
              >
                <Button icon={<CheckOutlined />} size="small" type="primary" />
              </Popconfirm>
            </Tooltip>
          )}
          
          {(record.status === 'FREE' || record.status === 'BUSY') && (
            <Tooltip title="Деактивировать">
              <Popconfirm
                title="Деактивировать курьера?"
                onConfirm={() => handleDeactivate(record.courierId)}
              >
                <Button icon={<StopOutlined />} size="small" danger />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  const couriers = data?.data || [];
  const totalCount = data?.total || 0;

  return (
    <div className="p-6">
      <Card
        title="Курьеры"
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => refetch()}
            >
              Обновить
            </Button>
            <Link href="/couriers/create">
              <Button type="primary" icon={<PlusOutlined />}>
                Добавить курьера
              </Button>
            </Link>
          </Space>
        }
      >
        {/* Filters */}
        <div className="mb-4 flex gap-4 flex-wrap">
          <Select
            mode="multiple"
            placeholder="Статус"
            style={{ minWidth: 200 }}
            value={statusFilter}
            onChange={setStatusFilter}
            options={[
              { value: 'FREE', label: 'Свободен' },
              { value: 'BUSY', label: 'Занят' },
              { value: 'UNAVAILABLE', label: 'Недоступен' },
              { value: 'ARCHIVED', label: 'В архиве' },
            ]}
            allowClear
          />
          <Select
            mode="multiple"
            placeholder="Транспорт"
            style={{ minWidth: 200 }}
            value={transportFilter}
            onChange={setTransportFilter}
            options={[
              { value: 'WALKING', label: 'Пешком' },
              { value: 'BICYCLE', label: 'Велосипед' },
              { value: 'MOTORCYCLE', label: 'Мотоцикл' },
              { value: 'CAR', label: 'Автомобиль' },
            ]}
            allowClear
          />
        </div>

        {/* Table */}
        <Table
          columns={columns}
          dataSource={couriers}
          rowKey="courierId"
          loading={isLoading}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: totalCount,
            showSizeChanger: true,
            showTotal: (total) => `Всего: ${total}`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
        />
      </Card>
    </div>
  );
}
