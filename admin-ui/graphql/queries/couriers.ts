import { gql } from '@apollo/client';

export const GET_COURIERS = gql`
  query GetCouriers(
    $filter: CourierFilterInput
    $pagination: PaginationInput
  ) {
    couriers(filter: $filter, pagination: $pagination) {
      couriers {
        courierId
        name
        phone
        email
        transportType
        maxDistanceKm
        status
        currentLoad
        maxLoad
        rating
        workZone
        successfulDeliveries
        failedDeliveries
        createdAt
        lastActiveAt
        workHours {
          startTime
          endTime
          workDays
        }
        currentLocation {
          latitude
          longitude
        }
      }
      totalCount
      pagination {
        currentPage
        pageSize
        totalPages
        totalItems
      }
    }
  }
`;

export const GET_COURIER = gql`
  query GetCourier($id: String!, $includeLocation: Boolean) {
    courier(id: $id, includeLocation: $includeLocation) {
      courierId
      name
      phone
      email
      transportType
      maxDistanceKm
      status
      currentLoad
      maxLoad
      rating
      workZone
      successfulDeliveries
      failedDeliveries
      createdAt
      lastActiveAt
      workHours {
        startTime
        endTime
        workDays
      }
      currentLocation {
        latitude
        longitude
      }
    }
  }
`;

export const GET_COURIER_DELIVERIES = gql`
  query GetCourierDeliveries($courierId: String!, $limit: Int) {
    courierDeliveries(courierId: $courierId, limit: $limit) {
      deliveries {
        packageId
        orderId
        status
        priority
        assignedAt
        deliveredAt
        pickupAddress {
          street
          city
          country
        }
        deliveryAddress {
          street
          city
          country
        }
      }
      totalCount
    }
  }
`;
