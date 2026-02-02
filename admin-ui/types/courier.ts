export type TransportType = 'UNSPECIFIED' | 'WALKING' | 'BICYCLE' | 'MOTORCYCLE' | 'CAR';

export type CourierStatus = 'UNSPECIFIED' | 'UNAVAILABLE' | 'FREE' | 'BUSY' | 'ARCHIVED';

export type PackageStatus = 
  | 'UNSPECIFIED' 
  | 'ACCEPTED' 
  | 'IN_POOL' 
  | 'ASSIGNED' 
  | 'IN_TRANSIT' 
  | 'DELIVERED' 
  | 'NOT_DELIVERED' 
  | 'REQUIRES_HANDLING';

export type Priority = 'UNSPECIFIED' | 'NORMAL' | 'URGENT';

export interface Location {
  latitude: number;
  longitude: number;
}

export interface WorkHours {
  startTime: string;
  endTime: string;
  workDays: number[];
}

export interface Address {
  street: string;
  city: string;
  postalCode?: string;
  country: string;
  latitude?: number;
  longitude?: number;
}

export interface Courier {
  courierId: string;
  name: string;
  phone: string;
  email: string;
  transportType: TransportType;
  maxDistanceKm: number;
  status: CourierStatus;
  currentLoad: number;
  maxLoad: number;
  rating: number;
  workHours?: WorkHours;
  workZone: string;
  currentLocation?: Location;
  successfulDeliveries: number;
  failedDeliveries: number;
  createdAt?: string;
  lastActiveAt?: string;
}

export interface DeliveryRecord {
  packageId: string;
  orderId: string;
  status: PackageStatus;
  pickupAddress?: Address;
  deliveryAddress?: Address;
  assignedAt?: string;
  deliveredAt?: string;
  priority: Priority;
}

export interface PaginationInfo {
  currentPage: number;
  pageSize: number;
  totalPages: number;
  totalItems: number;
}

export interface CourierListResponse {
  couriers: Courier[];
  totalCount: number;
  pagination?: PaginationInfo;
}

// Transport type labels
export const TRANSPORT_LABELS: Record<TransportType, string> = {
  UNSPECIFIED: 'Не указан',
  WALKING: 'Пешком',
  BICYCLE: 'Велосипед',
  MOTORCYCLE: 'Мотоцикл',
  CAR: 'Автомобиль',
};

// Status labels
export const STATUS_LABELS: Record<CourierStatus, string> = {
  UNSPECIFIED: 'Не указан',
  UNAVAILABLE: 'Недоступен',
  FREE: 'Свободен',
  BUSY: 'Занят',
  ARCHIVED: 'В архиве',
};

// Status colors for Ant Design Tag
export const STATUS_COLORS: Record<CourierStatus, string> = {
  UNSPECIFIED: 'default',
  UNAVAILABLE: 'default',
  FREE: 'success',
  BUSY: 'processing',
  ARCHIVED: 'error',
};
