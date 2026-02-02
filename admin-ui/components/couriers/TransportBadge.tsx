'use client';

import { Tag } from 'antd';
import { 
  CarOutlined, 
  ThunderboltOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { TransportType, TRANSPORT_LABELS } from '@/types/courier';

interface TransportBadgeProps {
  type: TransportType;
}

const TRANSPORT_ICONS: Record<TransportType, React.ReactNode> = {
  UNSPECIFIED: <UserOutlined />,
  WALKING: <UserOutlined />,
  BICYCLE: <ThunderboltOutlined />,
  MOTORCYCLE: <ThunderboltOutlined />,
  CAR: <CarOutlined />,
};

const TRANSPORT_COLORS: Record<TransportType, string> = {
  UNSPECIFIED: 'default',
  WALKING: 'green',
  BICYCLE: 'cyan',
  MOTORCYCLE: 'orange',
  CAR: 'blue',
};

export function TransportBadge({ type }: TransportBadgeProps) {
  return (
    <Tag icon={TRANSPORT_ICONS[type]} color={TRANSPORT_COLORS[type]}>
      {TRANSPORT_LABELS[type]}
    </Tag>
  );
}
