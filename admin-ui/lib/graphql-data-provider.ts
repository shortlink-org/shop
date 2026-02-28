'use client';

import type { CrudFilter, DataProvider, LogicalFilter } from '@refinedev/core';
import { apolloClient } from '@/lib/apollo-client';
import { GET_COURIERS, GET_COURIER } from '@/graphql/queries/couriers';
import {
  REGISTER_COURIER,
  ACTIVATE_COURIER,
  DEACTIVATE_COURIER,
  ARCHIVE_COURIER,
  UPDATE_COURIER_CONTACT,
  UPDATE_COURIER_SCHEDULE,
  CHANGE_COURIER_TRANSPORT,
} from '@/graphql/mutations/couriers';

function withId<T extends { courierId: string }>(item: T): T & { id: string } {
  return { ...item, id: item.courierId };
}

type CourierRecord = { courierId: string } & Record<string, unknown>;
type GetCouriersQueryResult = {
  couriers?: {
    couriers?: CourierRecord[];
    totalCount?: number;
  };
};
type GetCourierQueryResult = {
  courier?: CourierRecord | null;
};
type RegisterCourierMutationResult = {
  registerCourier?: {
    courierId: string;
    status?: string | null;
  } | null;
};

export const graphqlDataProvider: DataProvider = {
  getList: async ({ resource, pagination, filters }) => {
    if (resource !== 'couriers') {
      return { data: [], total: 0 } as any;
    }

    const page = pagination?.currentPage ?? 1;
    const pageSize = pagination?.pageSize ?? 20;

    const isFieldFilter = (filter: CrudFilter): filter is LogicalFilter =>
      filter.operator !== 'and' && filter.operator !== 'or';
    const fieldFilters = (filters ?? []).filter(isFieldFilter);
    const statusFilter = fieldFilters
      .filter((f) => f.field === 'status' && f.value)
      .flatMap((f) => (Array.isArray(f.value) ? f.value : [f.value]));
    const transportTypeFilter = fieldFilters
      .filter((f) => f.field === 'transportType' && f.value)
      .flatMap((f) => (Array.isArray(f.value) ? f.value : [f.value]));
    const zoneFilter = fieldFilters.find((f) => f.field === 'workZone' && f.value)?.value as string | undefined;

    const { data } = await apolloClient.query<GetCouriersQueryResult>({
      query: GET_COURIERS,
      variables: {
        filter: {
          ...(statusFilter?.length ? { statusFilter } : {}),
          ...(transportTypeFilter?.length ? { transportTypeFilter } : {}),
          ...(zoneFilter ? { zoneFilter } : {}),
        },
        pagination: { page, pageSize },
      },
    });

    const list = data?.couriers;
    if (!list) {
      return { data: [], total: 0 } as any;
    }

    const couriers = (list.couriers ?? []).map(withId);
    const total = list.totalCount ?? 0;

    return { data: couriers, total } as any;
  },

  getOne: async ({ resource, id }) => {
    if (resource !== 'couriers') {
      return { data: {} } as any;
    }

    const { data } = await apolloClient.query<GetCourierQueryResult>({
      query: GET_COURIER,
      variables: { id, includeLocation: true },
    });

    const courier = data?.courier;
    if (!courier) {
      return { data: {} } as any;
    }

    return { data: withId(courier) } as any;
  },

  create: async ({ resource, variables }: any) => {
    if (resource !== 'couriers') {
      return { data: {} } as any;
    }

    const workHours = variables?.workHours as { startTime: string; endTime: string; workDays: number[] } | undefined;
    if (!workHours?.startTime || !workHours?.endTime || !workHours?.workDays) {
      throw new Error('workHours (startTime, endTime, workDays) are required');
    }

    const { data } = await apolloClient.mutate<RegisterCourierMutationResult>({
      mutation: REGISTER_COURIER,
      variables: {
        input: {
          name: variables?.name,
          phone: variables?.phone,
          email: variables?.email,
          transportType: variables?.transportType,
          maxDistanceKm: Number(variables?.maxDistanceKm) ?? 10,
          workZone: variables?.workZone ?? '',
          workHours: {
            startTime: workHours.startTime,
            endTime: workHours.endTime,
            workDays: workHours.workDays,
          },
        },
      },
    });

    const result = data?.registerCourier;
    if (!result) {
      throw new Error('Failed to register courier');
    }

    return {
      data: withId({
        courierId: result.courierId,
        name: String(variables?.name),
        phone: String(variables?.phone),
        email: String(variables?.email),
        transportType: String(variables?.transportType),
        status: result.status ?? 'UNAVAILABLE',
        maxDistanceKm: Number(variables?.maxDistanceKm),
        workZone: String(variables?.workZone),
        workHours,
        currentLoad: 0,
        maxLoad: 0,
        rating: 0,
        successfulDeliveries: 0,
        failedDeliveries: 0,
      }),
    } as any;
  },

  update: async ({ resource, id, variables }: any) => {
    if (resource !== 'couriers') {
      return { data: {} } as any;
    }

    if (variables?.phone !== undefined || variables?.email !== undefined) {
      await apolloClient.mutate({
        mutation: UPDATE_COURIER_CONTACT,
        variables: {
          id,
          input: {
            ...(variables.phone !== undefined && { phone: variables.phone }),
            ...(variables.email !== undefined && { email: variables.email }),
          },
        },
      });
    }

    const workHours = variables?.workHours as { startTime: string; endTime: string; workDays: number[] } | undefined;
    const hasScheduleChanges =
      (workHours?.startTime && workHours?.endTime && workHours?.workDays) ||
      variables?.workZone !== undefined ||
      variables?.maxDistanceKm !== undefined;
    if (hasScheduleChanges) {
      await apolloClient.mutate({
        mutation: UPDATE_COURIER_SCHEDULE,
        variables: {
          id,
          input: {
            ...(workHours?.startTime && workHours?.endTime && workHours?.workDays && { workHours }),
            ...(variables?.workZone !== undefined && { workZone: variables.workZone }),
            ...(variables?.maxDistanceKm !== undefined && { maxDistanceKm: Number(variables.maxDistanceKm) }),
          },
        },
      });
    }

    if (variables?.transportType !== undefined) {
      await apolloClient.mutate({
        mutation: CHANGE_COURIER_TRANSPORT,
        variables: { id, transportType: variables.transportType },
      });
    }

    const { data } = await apolloClient.query<GetCourierQueryResult>({
      query: GET_COURIER,
      variables: { id, includeLocation: true },
    });

    const courier = data?.courier;
    if (!courier) {
      return { data: { id, courierId: id, ...variables } } as any;
    }

    return { data: withId({ ...courier, ...variables }) } as any;
  },

  deleteOne: async ({ resource, id }) => {
    if (resource !== 'couriers') {
      return { data: {} } as any;
    }

    await apolloClient.mutate({
      mutation: ARCHIVE_COURIER,
      variables: { id, reason: 'Archived from admin' },
    });

    return { data: { id, courierId: id } } as any;
  },

  getApiUrl: () => process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:9991/graphql',
};
