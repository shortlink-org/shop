import { gql } from '@apollo/client';

export const REGISTER_COURIER = gql`
  mutation RegisterCourier($input: RegisterCourierInput!) {
    registerCourier(input: $input) {
      courierId
      status
      createdAt
    }
  }
`;

export const ACTIVATE_COURIER = gql`
  mutation ActivateCourier($id: String!) {
    activateCourier(id: $id) {
      courierId
      status
    }
  }
`;

export const DEACTIVATE_COURIER = gql`
  mutation DeactivateCourier($id: String!, $reason: String) {
    deactivateCourier(id: $id, reason: $reason) {
      courierId
      status
    }
  }
`;

export const ARCHIVE_COURIER = gql`
  mutation ArchiveCourier($id: String!, $reason: String) {
    archiveCourier(id: $id, reason: $reason) {
      courierId
      status
    }
  }
`;

export const UPDATE_COURIER_CONTACT = gql`
  mutation UpdateCourierContact($id: String!, $input: UpdateContactInput!) {
    updateCourierContact(id: $id, input: $input) {
      courierId
      phone
      email
      updatedAt
    }
  }
`;

export const UPDATE_COURIER_SCHEDULE = gql`
  mutation UpdateCourierSchedule($id: String!, $input: UpdateScheduleInput!) {
    updateCourierSchedule(id: $id, input: $input) {
      courierId
      workZone
      maxDistanceKm
      workHours {
        startTime
        endTime
        workDays
      }
      updatedAt
    }
  }
`;

export const CHANGE_COURIER_TRANSPORT = gql`
  mutation ChangeCourierTransport($id: String!, $transportType: TransportType!) {
    changeCourierTransport(id: $id, transportType: $transportType) {
      courierId
      transportType
      maxLoad
      updatedAt
    }
  }
`;
