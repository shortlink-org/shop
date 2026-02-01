//! Proto <-> Domain type converters
//!
//! Conversion utilities between gRPC protobuf types and domain types.

use chrono::NaiveTime;
use tonic::Status;

use crate::domain::ports::CachedCourierState;
use crate::domain::model::courier::{Courier, CourierStatus as DomainCourierStatus, WorkHours as DomainWorkHours};
use crate::domain::model::vo::TransportType as DomainTransportType;

use super::{Courier as ProtoCourier, CourierStatus, TransportType, WorkHours as ProtoWorkHours};

/// Convert proto TransportType to domain
pub fn proto_to_domain_transport(t: TransportType) -> DomainTransportType {
    match t {
        TransportType::Walking => DomainTransportType::Walking,
        TransportType::Bicycle => DomainTransportType::Bicycle,
        TransportType::Motorcycle => DomainTransportType::Motorcycle,
        TransportType::Car => DomainTransportType::Car,
        TransportType::Unspecified => DomainTransportType::Walking,
    }
}

/// Convert domain TransportType to proto
pub fn domain_to_proto_transport(t: DomainTransportType) -> TransportType {
    match t {
        DomainTransportType::Walking => TransportType::Walking,
        DomainTransportType::Bicycle => TransportType::Bicycle,
        DomainTransportType::Motorcycle => TransportType::Motorcycle,
        DomainTransportType::Car => TransportType::Car,
    }
}

/// Convert domain CourierStatus to proto
pub fn domain_to_proto_status(s: DomainCourierStatus) -> CourierStatus {
    match s {
        DomainCourierStatus::Unavailable => CourierStatus::Unavailable,
        DomainCourierStatus::Free => CourierStatus::Free,
        DomainCourierStatus::Busy => CourierStatus::Busy,
        DomainCourierStatus::Archived => CourierStatus::Archived,
    }
}

/// Convert proto CourierStatus to domain
pub fn proto_to_domain_status(s: CourierStatus) -> Option<DomainCourierStatus> {
    match s {
        CourierStatus::Unavailable => Some(DomainCourierStatus::Unavailable),
        CourierStatus::Free => Some(DomainCourierStatus::Free),
        CourierStatus::Busy => Some(DomainCourierStatus::Busy),
        CourierStatus::Archived => Some(DomainCourierStatus::Archived),
        CourierStatus::Unspecified => None,
    }
}

/// Parse work hours from proto message
pub fn parse_work_hours(wh: Option<ProtoWorkHours>) -> Result<DomainWorkHours, Status> {
    let wh = wh.ok_or_else(|| Status::invalid_argument("work_hours is required"))?;

    let start = NaiveTime::parse_from_str(&wh.start_time, "%H:%M")
        .map_err(|_| Status::invalid_argument("invalid start_time format, use HH:MM"))?;
    let end = NaiveTime::parse_from_str(&wh.end_time, "%H:%M")
        .map_err(|_| Status::invalid_argument("invalid end_time format, use HH:MM"))?;

    let work_days: Vec<u8> = wh.work_days.iter().map(|&d| d as u8).collect();

    DomainWorkHours::new(start, end, work_days)
        .map_err(|e| Status::invalid_argument(format!("invalid work_hours: {}", e)))
}

/// Convert domain WorkHours to proto
pub fn domain_to_proto_work_hours(wh: &DomainWorkHours) -> ProtoWorkHours {
    ProtoWorkHours {
        start_time: wh.start.format("%H:%M").to_string(),
        end_time: wh.end.format("%H:%M").to_string(),
        work_days: wh.days.iter().map(|&d| d as i32).collect(),
    }
}

/// Convert domain Courier to proto Courier
pub fn courier_to_proto(courier: &Courier, state: Option<&CachedCourierState>) -> ProtoCourier {
    let created_at = courier.created_at();

    ProtoCourier {
        courier_id: courier.id().0.to_string(),
        name: courier.name().to_string(),
        phone: courier.phone().to_string(),
        email: courier.email().to_string(),
        transport_type: domain_to_proto_transport(courier.transport_type()).into(),
        max_distance_km: courier.max_distance_km(),
        status: state
            .map(|s| domain_to_proto_status(s.status))
            .unwrap_or(CourierStatus::Unavailable)
            .into(),
        current_load: state.map(|s| s.current_load as i32).unwrap_or(0),
        max_load: courier.max_load() as i32,
        rating: state.map(|s| s.rating).unwrap_or(0.0),
        work_hours: Some(domain_to_proto_work_hours(courier.work_hours())),
        work_zone: courier.work_zone().to_string(),
        current_location: None, // TODO: integrate with Geolocation Service
        successful_deliveries: state.map(|s| s.successful_deliveries as i32).unwrap_or(0),
        failed_deliveries: state.map(|s| s.failed_deliveries as i32).unwrap_or(0),
        created_at: Some(prost_types::Timestamp {
            seconds: created_at.timestamp(),
            nanos: created_at.timestamp_subsec_nanos() as i32,
        }),
        last_active_at: None,
    }
}

/// Create a timestamp for the current time
pub fn now_timestamp() -> prost_types::Timestamp {
    let now = chrono::Utc::now();
    prost_types::Timestamp {
        seconds: now.timestamp(),
        nanos: now.timestamp_subsec_nanos() as i32,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::courier::{Courier, WorkHours};
    use chrono::{NaiveTime, Utc};

    // ==================== Transport Type Conversion Tests ====================

    #[test]
    fn test_proto_to_domain_transport_walking() {
        assert_eq!(
            proto_to_domain_transport(TransportType::Walking),
            DomainTransportType::Walking
        );
    }

    #[test]
    fn test_proto_to_domain_transport_bicycle() {
        assert_eq!(
            proto_to_domain_transport(TransportType::Bicycle),
            DomainTransportType::Bicycle
        );
    }

    #[test]
    fn test_proto_to_domain_transport_motorcycle() {
        assert_eq!(
            proto_to_domain_transport(TransportType::Motorcycle),
            DomainTransportType::Motorcycle
        );
    }

    #[test]
    fn test_proto_to_domain_transport_car() {
        assert_eq!(
            proto_to_domain_transport(TransportType::Car),
            DomainTransportType::Car
        );
    }

    #[test]
    fn test_proto_to_domain_transport_unspecified_defaults_to_walking() {
        assert_eq!(
            proto_to_domain_transport(TransportType::Unspecified),
            DomainTransportType::Walking
        );
    }

    #[test]
    fn test_domain_to_proto_transport_all_variants() {
        assert_eq!(
            domain_to_proto_transport(DomainTransportType::Walking),
            TransportType::Walking
        );
        assert_eq!(
            domain_to_proto_transport(DomainTransportType::Bicycle),
            TransportType::Bicycle
        );
        assert_eq!(
            domain_to_proto_transport(DomainTransportType::Motorcycle),
            TransportType::Motorcycle
        );
        assert_eq!(
            domain_to_proto_transport(DomainTransportType::Car),
            TransportType::Car
        );
    }

    // ==================== Status Conversion Tests ====================

    #[test]
    fn test_domain_to_proto_status_unavailable() {
        assert_eq!(
            domain_to_proto_status(DomainCourierStatus::Unavailable),
            CourierStatus::Unavailable
        );
    }

    #[test]
    fn test_domain_to_proto_status_free() {
        assert_eq!(
            domain_to_proto_status(DomainCourierStatus::Free),
            CourierStatus::Free
        );
    }

    #[test]
    fn test_domain_to_proto_status_busy() {
        assert_eq!(
            domain_to_proto_status(DomainCourierStatus::Busy),
            CourierStatus::Busy
        );
    }

    #[test]
    fn test_domain_to_proto_status_archived() {
        assert_eq!(
            domain_to_proto_status(DomainCourierStatus::Archived),
            CourierStatus::Archived
        );
    }

    #[test]
    fn test_proto_to_domain_status_all_variants() {
        assert_eq!(
            proto_to_domain_status(CourierStatus::Unavailable),
            Some(DomainCourierStatus::Unavailable)
        );
        assert_eq!(
            proto_to_domain_status(CourierStatus::Free),
            Some(DomainCourierStatus::Free)
        );
        assert_eq!(
            proto_to_domain_status(CourierStatus::Busy),
            Some(DomainCourierStatus::Busy)
        );
        assert_eq!(
            proto_to_domain_status(CourierStatus::Archived),
            Some(DomainCourierStatus::Archived)
        );
    }

    #[test]
    fn test_proto_to_domain_status_unspecified_returns_none() {
        assert_eq!(proto_to_domain_status(CourierStatus::Unspecified), None);
    }

    // ==================== Work Hours Conversion Tests ====================

    #[test]
    fn test_parse_work_hours_valid() {
        let proto_wh = ProtoWorkHours {
            start_time: "09:00".to_string(),
            end_time: "18:00".to_string(),
            work_days: vec![1, 2, 3, 4, 5],
        };

        let result = parse_work_hours(Some(proto_wh));
        assert!(result.is_ok());

        let wh = result.unwrap();
        assert_eq!(wh.start, NaiveTime::from_hms_opt(9, 0, 0).unwrap());
        assert_eq!(wh.end, NaiveTime::from_hms_opt(18, 0, 0).unwrap());
        assert_eq!(wh.days, vec![1, 2, 3, 4, 5]);
    }

    #[test]
    fn test_parse_work_hours_none_returns_error() {
        let result = parse_work_hours(None);
        assert!(result.is_err());
        let status = result.unwrap_err();
        assert_eq!(status.code(), tonic::Code::InvalidArgument);
    }

    #[test]
    fn test_parse_work_hours_invalid_start_time_format() {
        let proto_wh = ProtoWorkHours {
            start_time: "9am".to_string(), // Invalid format
            end_time: "18:00".to_string(),
            work_days: vec![1, 2, 3],
        };

        let result = parse_work_hours(Some(proto_wh));
        assert!(result.is_err());
    }

    #[test]
    fn test_parse_work_hours_invalid_end_time_format() {
        let proto_wh = ProtoWorkHours {
            start_time: "09:00".to_string(),
            end_time: "6pm".to_string(), // Invalid format
            work_days: vec![1, 2, 3],
        };

        let result = parse_work_hours(Some(proto_wh));
        assert!(result.is_err());
    }

    #[test]
    fn test_domain_to_proto_work_hours() {
        let domain_wh = DomainWorkHours::new(
            NaiveTime::from_hms_opt(8, 30, 0).unwrap(),
            NaiveTime::from_hms_opt(17, 45, 0).unwrap(),
            vec![1, 2, 3, 4, 5, 6],
        )
        .unwrap();

        let proto_wh = domain_to_proto_work_hours(&domain_wh);

        assert_eq!(proto_wh.start_time, "08:30");
        assert_eq!(proto_wh.end_time, "17:45");
        assert_eq!(proto_wh.work_days, vec![1, 2, 3, 4, 5, 6]);
    }

    // ==================== Timestamp Tests ====================

    #[test]
    fn test_now_timestamp_returns_valid_timestamp() {
        let before = Utc::now().timestamp();
        let ts = now_timestamp();
        let after = Utc::now().timestamp();

        // Timestamp should be between before and after
        assert!(ts.seconds >= before);
        assert!(ts.seconds <= after);
        // Nanos should be positive
        assert!(ts.nanos >= 0);
    }

    // ==================== Courier Conversion Tests ====================

    fn create_test_courier() -> Courier {
        let work_hours = WorkHours::new(
            NaiveTime::from_hms_opt(9, 0, 0).unwrap(),
            NaiveTime::from_hms_opt(18, 0, 0).unwrap(),
            vec![1, 2, 3, 4, 5],
        )
        .unwrap();

        Courier::builder(
            "Test Courier".to_string(),
            "+49123456789".to_string(),
            "test@example.com".to_string(),
            DomainTransportType::Bicycle,
            10.0,
            "Berlin-Mitte".to_string(),
            work_hours,
        )
        .build()
        .unwrap()
    }

    #[test]
    fn test_courier_to_proto_without_state() {
        let courier = create_test_courier();

        let proto = courier_to_proto(&courier, None);

        assert_eq!(proto.name, "Test Courier");
        assert_eq!(proto.phone, "+49123456789");
        assert_eq!(proto.email, "test@example.com");
        assert_eq!(proto.transport_type, TransportType::Bicycle as i32);
        assert_eq!(proto.work_zone, "Berlin-Mitte");
        // Without state, defaults to Unavailable
        assert_eq!(proto.status, CourierStatus::Unavailable as i32);
        assert_eq!(proto.current_load, 0);
        assert_eq!(proto.rating, 0.0);
    }

    #[test]
    fn test_courier_to_proto_with_state() {
        let courier = create_test_courier();
        let state = CachedCourierState {
            status: DomainCourierStatus::Free,
            current_load: 1,
            max_load: 2,
            rating: 4.5,
            successful_deliveries: 50,
            failed_deliveries: 2,
        };

        let proto = courier_to_proto(&courier, Some(&state));

        assert_eq!(proto.status, CourierStatus::Free as i32);
        assert_eq!(proto.current_load, 1);
        assert_eq!(proto.rating, 4.5);
        assert_eq!(proto.successful_deliveries, 50);
        assert_eq!(proto.failed_deliveries, 2);
    }

    #[test]
    fn test_courier_to_proto_includes_work_hours() {
        let courier = create_test_courier();

        let proto = courier_to_proto(&courier, None);

        assert!(proto.work_hours.is_some());
        let wh = proto.work_hours.unwrap();
        assert_eq!(wh.start_time, "09:00");
        assert_eq!(wh.end_time, "18:00");
        assert_eq!(wh.work_days, vec![1, 2, 3, 4, 5]);
    }

    #[test]
    fn test_courier_to_proto_includes_timestamp() {
        let courier = create_test_courier();

        let proto = courier_to_proto(&courier, None);

        assert!(proto.created_at.is_some());
        let created_at = proto.created_at.unwrap();
        assert!(created_at.seconds > 0);
    }
}
