export type OrderItem = {
  id?: string | null;
  quantity?: number | null;
  price?: number | null;
};

export type DeliveryAddress = {
  street?: string | null;
  city?: string | null;
  country?: string | null;
  latitude?: number | null;
  longitude?: number | null;
};

export type DeliveryPeriod = {
  startTime?: string | null;
  endTime?: string | null;
};

export type RecipientContacts = {
  recipientName?: string | null;
  recipientPhone?: string | null;
  recipientEmail?: string | null;
};

export type DeliveryInfo = {
  pickupAddress?: DeliveryAddress | null;
  deliveryAddress?: DeliveryAddress | null;
  deliveryPeriod?: DeliveryPeriod | null;
  recipientContacts?: RecipientContacts | null;
  priority?: string | null;
};

export type OrderState = {
  id?: string | null;
  status?: string | null;
  items?: OrderItem[] | null;
  deliveryInfo?: DeliveryInfo | null;
};

export type DeliveryTrackingLocation = {
  latitude?: number | null;
  longitude?: number | null;
};

export type DeliveryTrackingCourier = {
  courierId?: string | null;
  name?: string | null;
  phone?: string | null;
  transportType?: string | null;
  status?: string | null;
  lastActiveAt?: string | null;
  currentLocation?: DeliveryTrackingLocation | null;
};

export type DeliveryTracking = {
  orderId?: string | null;
  packageId?: string | null;
  status?: string | null;
  estimatedMinutesRemaining?: number | null;
  distanceKmRemaining?: number | null;
  estimatedArrivalAt?: string | null;
  assignedAt?: string | null;
  deliveredAt?: string | null;
  courier?: DeliveryTrackingCourier | null;
};
