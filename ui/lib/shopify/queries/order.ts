export const getOrderTrackingPageQuery = /* GraphQL */ `
  query GetOrderTrackingPage($id: String!) {
    getOrder(id: $id) {
      order {
        id
        status
        deliveryInfo {
          deliveryAddress {
            street
            city
            postalCode
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

export const getDeliveryTrackingQuery = /* GraphQL */ `
  query GetDeliveryTracking($id: String!) {
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
