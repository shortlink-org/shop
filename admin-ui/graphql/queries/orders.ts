import { gql } from '@apollo/client';

export const GET_ORDER_LOOKUP = gql`
  query GetOrderLookup($id: String!) {
    getOrder(id: $id) {
      order {
        id
        status
        items {
          id
          quantity
          price
        }
        deliveryInfo {
          priority
          pickupAddress {
            street
            city
            country
            latitude
            longitude
          }
          deliveryAddress {
            street
            city
            country
            latitude
            longitude
          }
          deliveryPeriod {
            startTime
            endTime
          }
          recipientContacts {
            recipientName
            recipientPhone
            recipientEmail
          }
        }
      }
    }
    deliveryTracking(orderId: $id) {
      orderId
      packageId
      status
      estimatedMinutesRemaining
      distanceKmRemaining
      estimatedArrivalAt
      assignedAt
      deliveredAt
      courier {
        courierId
        name
        phone
        transportType
        status
        lastActiveAt
        currentLocation {
          latitude
          longitude
        }
      }
    }
  }
`;
