'use client';

import { Tag } from 'antd';
import { CourierStatus, STATUS_LABELS, STATUS_COLORS } from '@/types/courier';

interface CourierStatusBadgeProps {
  status: CourierStatus;
}

export function CourierStatusBadge({ status }: CourierStatusBadgeProps) {
  return (
    <Tag color={STATUS_COLORS[status]}>
      {STATUS_LABELS[status]}
    </Tag>
  );
}
