'use client';

import { useParams } from 'next/navigation';
import { useOne } from '@refinedev/core';
import { useMutation } from '@apollo/client/react';
import { 
  Card, 
  Descriptions, 
  Button, 
  Space, 
  Spin, 
  Rate,
  message,
  Popconfirm,
  Divider,
} from 'antd';
import { 
  ArrowLeftOutlined,
  CheckOutlined,
  StopOutlined,
  DeleteOutlined,
  EditOutlined,
  PhoneOutlined,
  MailOutlined,
  EnvironmentOutlined,
} from '@ant-design/icons';
import Link from 'next/link';

import { CourierStatusBadge } from '@/components/couriers/CourierStatusBadge';
import { TransportBadge } from '@/components/couriers/TransportBadge';
import { ACTIVATE_COURIER, DEACTIVATE_COURIER, ARCHIVE_COURIER } from '@/graphql/mutations/couriers';
import type { Courier } from '@/types/courier';

const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

export default function CourierDetailPage() {
  const params = useParams();
  const courierId = params.id as string;

  const { query } = useOne<Courier>({
    resource: 'couriers',
    id: courierId,
  });
  
  const { data, isLoading, refetch } = query;

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
  const [archiveCourier] = useMutation(ARCHIVE_COURIER, {
    onCompleted: () => {
      message.success('Courier archived');
      refetch();
    },
    onError: (e) => message.error(e.message),
  });

  const handleActivate = () => {
    activateCourier({ variables: { id: courierId } });
  };

  const handleDeactivate = () => {
    deactivateCourier({ variables: { id: courierId } });
  };

  const handleArchive = () => {
    archiveCourier({ variables: { id: courierId } });
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  const courier = data?.data as Courier | undefined;

  if (!courier) {
    return (
      <div className="p-6">
        <Card>
          <p>Courier not found</p>
          <Link href="/couriers">
            <Button icon={<ArrowLeftOutlined />}>Back to list</Button>
          </Link>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-4 flex justify-between items-center flex-wrap gap-4">
        <Space>
          <Link href="/couriers">
            <Button icon={<ArrowLeftOutlined />}>Back</Button>
          </Link>
          <h1 className="text-2xl font-bold m-0">{courier.name}</h1>
          <CourierStatusBadge status={courier.status} />
        </Space>
        
        <Space wrap>
          {courier.status === 'UNAVAILABLE' && (
            <Popconfirm
              title="Activate courier?"
              onConfirm={handleActivate}
            >
              <Button icon={<CheckOutlined />} type="primary">
                Activate
              </Button>
            </Popconfirm>
          )}
          
          {(courier.status === 'FREE' || courier.status === 'BUSY') && (
            <Popconfirm
              title="Deactivate courier?"
              onConfirm={handleDeactivate}
            >
              <Button icon={<StopOutlined />}>
                Deactivate
              </Button>
            </Popconfirm>
          )}
          
          {courier.status !== 'ARCHIVED' && (
            <Popconfirm
              title="Archive courier? This action cannot be undone."
              onConfirm={handleArchive}
              okButtonProps={{ danger: true }}
            >
              <Button icon={<DeleteOutlined />} danger>
                Archive
              </Button>
            </Popconfirm>
          )}
          
          <Link href={`/couriers/${courierId}/edit`}>
            <Button icon={<EditOutlined />}>Edit</Button>
          </Link>
        </Space>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Basic Info */}
        <Card title="Basic information">
          <Descriptions column={1}>
            <Descriptions.Item label="ID">{courier.courierId}</Descriptions.Item>
            <Descriptions.Item label="Name">{courier.name}</Descriptions.Item>
            <Descriptions.Item label={<><PhoneOutlined /> Phone</>}>
              <a href={`tel:${courier.phone}`}>{courier.phone}</a>
            </Descriptions.Item>
            <Descriptions.Item label={<><MailOutlined /> Email</>}>
              <a href={`mailto:${courier.email}`}>{courier.email}</a>
            </Descriptions.Item>
            <Descriptions.Item label="Rating">
              <Rate disabled value={courier.rating} allowHalf />
              <span className="ml-2">({courier.rating?.toFixed(1)})</span>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Transport & Work */}
        <Card title="Transport and work">
          <Descriptions column={1}>
            <Descriptions.Item label="Transport">
              <TransportBadge type={courier.transportType} />
            </Descriptions.Item>
            <Descriptions.Item label="Max distance">
              {courier.maxDistanceKm} km
            </Descriptions.Item>
            <Descriptions.Item label="Load">
              {courier.currentLoad} / {courier.maxLoad} packages
            </Descriptions.Item>
            <Descriptions.Item label={<><EnvironmentOutlined /> Zone</>}>
              {courier.workZone}
            </Descriptions.Item>
            {courier.workHours && (
              <>
                <Descriptions.Item label="Working hours">
                  {courier.workHours.startTime} - {courier.workHours.endTime}
                </Descriptions.Item>
                <Descriptions.Item label="Working days">
                  {courier.workHours.workDays?.map((d: number) => WEEKDAYS[d]).join(', ')}
                </Descriptions.Item>
              </>
            )}
          </Descriptions>
        </Card>

        {/* Statistics */}
        <Card title="Statistics">
          <Descriptions column={2}>
            <Descriptions.Item label="Successful deliveries">
              <span className="text-green-600 font-bold">
                {courier.successfulDeliveries}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Failed deliveries">
              <span className="text-red-600 font-bold">
                {courier.failedDeliveries}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Registration date">
              {courier.createdAt 
                ? new Date(courier.createdAt).toLocaleDateString('en-US') 
                : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Last activity">
              {courier.lastActiveAt 
                ? new Date(courier.lastActiveAt).toLocaleString('en-US') 
                : '—'}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Location */}
        {courier.currentLocation && (
          <Card title="Current location">
            <Descriptions column={2}>
              <Descriptions.Item label="Latitude">
                {courier.currentLocation.latitude?.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="Longitude">
                {courier.currentLocation.longitude?.toFixed(6)}
              </Descriptions.Item>
            </Descriptions>
            <div className="mt-4 p-4 bg-gray-100 rounded text-center text-gray-500">
              Map will be added later
            </div>
          </Card>
        )}
      </div>

      {/* Recent Deliveries Placeholder */}
      <Divider />
      <Card title="Recent deliveries" className="mt-6">
        <div className="text-center text-gray-500 py-8">
          Delivery data will be loaded after the GraphQL API is connected
        </div>
      </Card>
    </div>
  );
}
