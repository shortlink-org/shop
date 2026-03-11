'use client';

import { useState } from 'react';
import { useList } from '@refinedev/core';
import { useMutation } from '@apollo/client/react';
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
import { ACTIVATE_COURIER, DEACTIVATE_COURIER } from '@/graphql/mutations/couriers';
import type { Courier, CourierStatus, TransportType } from '@/types/courier';

export default function CouriersListPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState<CourierStatus[]>([]);
  const [transportFilter, setTransportFilter] = useState<TransportType[]>([]);

  const filters = [
    ...(statusFilter.length ? [{ field: 'status' as const, operator: 'in' as const, value: statusFilter }] : []),
    ...(transportFilter.length ? [{ field: 'transportType' as const, operator: 'in' as const, value: transportFilter }] : []),
  ];

  const { query, result } = useList<Courier>({
    resource: 'couriers',
    pagination: { currentPage: page, pageSize },
    filters,
  });
  
  const { isLoading, refetch } = query;
  const data = result;

  const [activateCourier] = useMutation(ACTIVATE_COURIER, {
    onCompleted: () => {
      message.success('Courier activated');
      refetch();
    },
    onError: (e) => message.error(e.message),
  });
  const [deactivateCourier] = useMutation(DEACTIVATE_COURIER, {
    onCompleted: () => {
      message.success('Courier deactivated');
      refetch();
    },
    onError: (e) => message.error(e.message),
  });

  const handleActivate = (id: string) => {
    activateCourier({ variables: { id } });
  };

  const handleDeactivate = (id: string) => {
    deactivateCourier({ variables: { id } });
  };

  const columns: ColumnsType<Courier> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Courier) => (
        <Link href={`/couriers/${record.courierId}`}>
          <span className="text-blue-600 hover:underline">{name}</span>
        </Link>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: CourierStatus) => <CourierStatusBadge status={status} />,
    },
    {
      title: 'Transport',
      dataIndex: 'transportType',
      key: 'transportType',
      render: (type: TransportType) => <TransportBadge type={type} />,
    },
    {
      title: 'Zone',
      dataIndex: 'workZone',
      key: 'workZone',
    },
    {
      title: 'Rating',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating: number) => (
        <Rate disabled defaultValue={rating} allowHalf />
      ),
    },
    {
      title: 'Load',
      key: 'load',
      render: (_: unknown, record: Courier) => (
        <span>
          {record.currentLoad} / {record.maxLoad}
        </span>
      ),
    },
    {
      title: 'Deliveries',
      key: 'deliveries',
      render: (_: unknown, record: Courier) => (
        <Tooltip title={`Successful: ${record.successfulDeliveries}, Failed: ${record.failedDeliveries}`}>
          <span className="text-green-600">{record.successfulDeliveries}</span>
          {' / '}
          <span className="text-red-600">{record.failedDeliveries}</span>
        </Tooltip>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, record: Courier) => (
        <Space size="small">
          <Tooltip title="View">
            <Link href={`/couriers/${record.courierId}`}>
              <Button icon={<EyeOutlined />} size="small" />
            </Link>
          </Tooltip>
          
          {record.status === 'UNAVAILABLE' && (
            <Tooltip title="Activate">
              <Popconfirm
                title="Activate courier?"
                onConfirm={() => handleActivate(record.courierId)}
              >
                <Button icon={<CheckOutlined />} size="small" type="primary" />
              </Popconfirm>
            </Tooltip>
          )}
          
          {(record.status === 'FREE' || record.status === 'BUSY') && (
            <Tooltip title="Deactivate">
              <Popconfirm
                title="Deactivate courier?"
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
        title="Couriers"
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => refetch()}
            >
              Refresh
            </Button>
            <Link href="/couriers/create">
              <Button type="primary" icon={<PlusOutlined />}>
                Add courier
              </Button>
            </Link>
          </Space>
        }
      >
        {/* Filters */}
        <div className="mb-4 flex gap-4 flex-wrap">
          <Select
            mode="multiple"
            placeholder="Status"
            style={{ minWidth: 200 }}
            value={statusFilter}
            onChange={setStatusFilter}
            options={[
              { value: 'FREE', label: 'Available' },
              { value: 'BUSY', label: 'Busy' },
              { value: 'UNAVAILABLE', label: 'Unavailable' },
              { value: 'ARCHIVED', label: 'Archived' },
            ]}
            allowClear
          />
          <Select
            mode="multiple"
            placeholder="Transport"
            style={{ minWidth: 200 }}
            value={transportFilter}
            onChange={setTransportFilter}
            options={[
              { value: 'WALKING', label: 'Walking' },
              { value: 'BICYCLE', label: 'Bicycle' },
              { value: 'MOTORCYCLE', label: 'Motorcycle' },
              { value: 'CAR', label: 'Car' },
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
            showTotal: (total) => `Total: ${total}`,
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
