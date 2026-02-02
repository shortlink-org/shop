'use client';

import { useParams } from 'next/navigation';
import { useOne } from '@refinedev/core';
import { 
  Card, 
  Descriptions, 
  Button, 
  Space, 
  Spin, 
  Rate,
  Tag,
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
import type { Courier } from '@/types/courier';

const WEEKDAYS = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

export default function CourierDetailPage() {
  const params = useParams();
  const courierId = params.id as string;

  const { query } = useOne<Courier>({
    resource: 'couriers',
    id: courierId,
  });
  
  const { data, isLoading, refetch } = query;

  const handleActivate = () => {
    message.success('Курьер активирован (mock)');
    refetch();
  };

  const handleDeactivate = () => {
    message.success('Курьер деактивирован (mock)');
    refetch();
  };

  const handleArchive = () => {
    message.success('Курьер архивирован (mock)');
    refetch();
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
          <p>Курьер не найден</p>
          <Link href="/couriers">
            <Button icon={<ArrowLeftOutlined />}>Назад к списку</Button>
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
            <Button icon={<ArrowLeftOutlined />}>Назад</Button>
          </Link>
          <h1 className="text-2xl font-bold m-0">{courier.name}</h1>
          <CourierStatusBadge status={courier.status} />
        </Space>
        
        <Space wrap>
          {courier.status === 'UNAVAILABLE' && (
            <Popconfirm
              title="Активировать курьера?"
              onConfirm={handleActivate}
            >
              <Button icon={<CheckOutlined />} type="primary">
                Активировать
              </Button>
            </Popconfirm>
          )}
          
          {(courier.status === 'FREE' || courier.status === 'BUSY') && (
            <Popconfirm
              title="Деактивировать курьера?"
              onConfirm={handleDeactivate}
            >
              <Button icon={<StopOutlined />}>
                Деактивировать
              </Button>
            </Popconfirm>
          )}
          
          {courier.status !== 'ARCHIVED' && (
            <Popconfirm
              title="Архивировать курьера? Это действие нельзя отменить."
              onConfirm={handleArchive}
              okButtonProps={{ danger: true }}
            >
              <Button icon={<DeleteOutlined />} danger>
                Архивировать
              </Button>
            </Popconfirm>
          )}
          
          <Link href={`/couriers/${courierId}/edit`}>
            <Button icon={<EditOutlined />}>Редактировать</Button>
          </Link>
        </Space>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Basic Info */}
        <Card title="Основная информация">
          <Descriptions column={1}>
            <Descriptions.Item label="ID">{courier.courierId}</Descriptions.Item>
            <Descriptions.Item label="Имя">{courier.name}</Descriptions.Item>
            <Descriptions.Item label={<><PhoneOutlined /> Телефон</>}>
              <a href={`tel:${courier.phone}`}>{courier.phone}</a>
            </Descriptions.Item>
            <Descriptions.Item label={<><MailOutlined /> Email</>}>
              <a href={`mailto:${courier.email}`}>{courier.email}</a>
            </Descriptions.Item>
            <Descriptions.Item label="Рейтинг">
              <Rate disabled value={courier.rating} allowHalf />
              <span className="ml-2">({courier.rating?.toFixed(1)})</span>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Transport & Work */}
        <Card title="Транспорт и работа">
          <Descriptions column={1}>
            <Descriptions.Item label="Транспорт">
              <TransportBadge type={courier.transportType} />
            </Descriptions.Item>
            <Descriptions.Item label="Макс. дистанция">
              {courier.maxDistanceKm} км
            </Descriptions.Item>
            <Descriptions.Item label="Загрузка">
              {courier.currentLoad} / {courier.maxLoad} посылок
            </Descriptions.Item>
            <Descriptions.Item label={<><EnvironmentOutlined /> Зона</>}>
              {courier.workZone}
            </Descriptions.Item>
            {courier.workHours && (
              <>
                <Descriptions.Item label="Рабочие часы">
                  {courier.workHours.startTime} - {courier.workHours.endTime}
                </Descriptions.Item>
                <Descriptions.Item label="Рабочие дни">
                  {courier.workHours.workDays?.map((d: number) => WEEKDAYS[d]).join(', ')}
                </Descriptions.Item>
              </>
            )}
          </Descriptions>
        </Card>

        {/* Statistics */}
        <Card title="Статистика">
          <Descriptions column={2}>
            <Descriptions.Item label="Успешных доставок">
              <span className="text-green-600 font-bold">
                {courier.successfulDeliveries}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Неудачных доставок">
              <span className="text-red-600 font-bold">
                {courier.failedDeliveries}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="Дата регистрации">
              {courier.createdAt 
                ? new Date(courier.createdAt).toLocaleDateString('ru-RU') 
                : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Последняя активность">
              {courier.lastActiveAt 
                ? new Date(courier.lastActiveAt).toLocaleString('ru-RU') 
                : '—'}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Location */}
        {courier.currentLocation && (
          <Card title="Текущая локация">
            <Descriptions column={2}>
              <Descriptions.Item label="Широта">
                {courier.currentLocation.latitude?.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="Долгота">
                {courier.currentLocation.longitude?.toFixed(6)}
              </Descriptions.Item>
            </Descriptions>
            <div className="mt-4 p-4 bg-gray-100 rounded text-center text-gray-500">
              Карта будет добавлена позже
            </div>
          </Card>
        )}
      </div>

      {/* Recent Deliveries Placeholder */}
      <Divider />
      <Card title="Недавние доставки" className="mt-6">
        <div className="text-center text-gray-500 py-8">
          Данные о доставках будут загружены после подключения к GraphQL API
        </div>
      </Card>
    </div>
  );
}
