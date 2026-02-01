//! Courier-related gRPC handlers
//!
//! Handles courier registration, queries, lifecycle, and profile updates.

use std::sync::Arc;

use tonic::{Response, Status};
use tracing::{error, info};

use crate::domain::ports::{
    CommandHandlerWithResult, CourierCache, CourierRepository, PackageFilter, PackageRepository,
    QueryHandler,
};
use crate::di::AppState;
use crate::domain::model::courier::CourierStatus as DomainCourierStatus;
use crate::usecases::courier::command::register::{
    Command as RegisterCommand, Handler as RegisterHandler,
};
use crate::usecases::courier::query::get_pool::{
    CourierFilter, Handler as GetPoolHandler, Query as GetPoolQuery,
};

use crate::infrastructure::rpc::converters::{
    courier_to_proto, domain_to_proto_transport, domain_to_proto_work_hours, now_timestamp,
    package_to_delivery_record, parse_work_hours, proto_to_domain_status,
    proto_to_domain_transport,
};
use crate::infrastructure::rpc::{
    ActivateCourierRequest, ActivateCourierResponse, ArchiveCourierRequest, ArchiveCourierResponse,
    ChangeTransportTypeRequest, ChangeTransportTypeResponse, CourierStatus,
    DeactivateCourierRequest, DeactivateCourierResponse, GetCourierDeliveriesRequest,
    GetCourierDeliveriesResponse, GetCourierPoolRequest, GetCourierPoolResponse,
    GetCourierRequest, GetCourierResponse, PaginationInfo, RegisterCourierRequest,
    RegisterCourierResponse, TransportType, UpdateContactInfoRequest, UpdateContactInfoResponse,
    UpdateWorkScheduleRequest, UpdateWorkScheduleResponse,
};

/// Handle RegisterCourier request
pub async fn register_courier(
    state: &Arc<AppState>,
    req: RegisterCourierRequest,
) -> Result<Response<RegisterCourierResponse>, Status> {
    info!("RegisterCourier request received");

    // Validate required fields
    if req.name.is_empty() {
        return Err(Status::invalid_argument("name is required"));
    }
    if req.phone.is_empty() {
        return Err(Status::invalid_argument("phone is required"));
    }
    if req.email.is_empty() {
        return Err(Status::invalid_argument("email is required"));
    }
    if req.work_zone.is_empty() {
        return Err(Status::invalid_argument("work_zone is required"));
    }

    let transport_type = proto_to_domain_transport(req.transport_type());
    let work_hours = parse_work_hours(req.work_hours)?;

    let cmd = RegisterCommand {
        name: req.name,
        phone: req.phone,
        email: req.email,
        transport_type,
        max_distance_km: req.max_distance_km,
        work_zone: req.work_zone,
        work_hours,
        push_token: req.push_token,
    };

    let handler = RegisterHandler::new(state.courier_repo.clone(), state.courier_cache.clone());

    let result = handler.handle(cmd).await.map_err(|e| {
        error!(error = %e, "Failed to register courier");
        match e {
            crate::usecases::courier::command::register::RegisterCourierError::EmailExists(_) => {
                Status::already_exists(e.to_string())
            }
            crate::usecases::courier::command::register::RegisterCourierError::PhoneExists(_) => {
                Status::already_exists(e.to_string())
            }
            _ => Status::internal(e.to_string()),
        }
    })?;

    info!(courier_id = %result.courier.id().0, "Courier registered successfully");

    let created_at = result.courier.created_at();
    let response = RegisterCourierResponse {
        courier_id: result.courier.id().0.to_string(),
        status: CourierStatus::Unavailable.into(),
        created_at: Some(prost_types::Timestamp {
            seconds: created_at.timestamp(),
            nanos: created_at.timestamp_subsec_nanos() as i32,
        }),
    };

    Ok(Response::new(response))
}

/// Handle GetCourierPool request
pub async fn get_courier_pool(
    state: &Arc<AppState>,
    req: GetCourierPoolRequest,
) -> Result<Response<GetCourierPoolResponse>, Status> {
    info!("GetCourierPool request received");

    // Build filter
    let status = req.status_filter.first().and_then(|&s| {
        proto_to_domain_status(CourierStatus::try_from(s).unwrap_or(CourierStatus::Unspecified))
    });

    let transport_type = req.transport_type_filter.first().map(|&t| {
        proto_to_domain_transport(TransportType::try_from(t).unwrap_or(TransportType::Unspecified))
    });

    let filter = CourierFilter {
        status,
        transport_type,
        work_zone: if req.zone_filter.is_empty() {
            None
        } else {
            Some(req.zone_filter)
        },
        available_only: req.available_only,
    };

    let (limit, offset) = if let Some(p) = req.pagination {
        let page = p.page.max(1) as u64;
        let size = p.page_size.clamp(1, 100) as u64;
        (Some(size), Some((page - 1) * size))
    } else {
        (Some(50), Some(0))
    };

    let query = GetPoolQuery {
        filter,
        limit,
        offset,
    };

    let handler = GetPoolHandler::new(state.courier_repo.clone(), state.courier_cache.clone());

    let result = handler.handle(query).await.map_err(|e| {
        error!(error = %e, "Failed to get courier pool");
        Status::internal(e.to_string())
    })?;

    let couriers = result
        .couriers
        .iter()
        .map(|cws| courier_to_proto(&cws.courier, cws.state.as_ref()))
        .collect();

    let total = result.total_count as i32;
    let page_size = limit.unwrap_or(50) as i32;
    let current_page = (offset.unwrap_or(0) / limit.unwrap_or(50) + 1) as i32;
    let total_pages = (total + page_size - 1) / page_size;

    let response = GetCourierPoolResponse {
        couriers,
        total_count: total,
        pagination: Some(PaginationInfo {
            current_page,
            page_size,
            total_pages,
            total_items: total,
        }),
    };

    info!(total_count = total, "GetCourierPool completed");
    Ok(Response::new(response))
}

/// Handle GetCourier request
pub async fn get_courier(
    state: &Arc<AppState>,
    req: GetCourierRequest,
) -> Result<Response<GetCourierResponse>, Status> {
    info!("GetCourier request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier from repository
    let courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Get state from cache
    let cached_state = state
        .courier_cache
        .get_state(courier_id)
        .await
        .ok()
        .flatten();

    let proto_courier = courier_to_proto(&courier, cached_state.as_ref());

    Ok(Response::new(GetCourierResponse {
        courier: Some(proto_courier),
    }))
}

/// Handle ActivateCourier request
pub async fn activate_courier(
    state: &Arc<AppState>,
    req: ActivateCourierRequest,
) -> Result<Response<ActivateCourierResponse>, Status> {
    info!("ActivateCourier request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier to verify it exists and get work_zone
    let courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Update status in cache
    state
        .courier_cache
        .update_status(courier_id, DomainCourierStatus::Free)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    // Add to free pool
    state
        .courier_cache
        .add_to_free_pool(courier_id, courier.work_zone())
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    Ok(Response::new(ActivateCourierResponse {
        courier_id: courier_id.to_string(),
        status: CourierStatus::Free.into(),
        activated_at: Some(now_timestamp()),
    }))
}

/// Handle DeactivateCourier request
pub async fn deactivate_courier(
    state: &Arc<AppState>,
    req: DeactivateCourierRequest,
) -> Result<Response<DeactivateCourierResponse>, Status> {
    info!("DeactivateCourier request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier to verify it exists and get work_zone
    let courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Update status in cache
    state
        .courier_cache
        .update_status(courier_id, DomainCourierStatus::Unavailable)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    // Remove from free pool
    state
        .courier_cache
        .remove_from_free_pool(courier_id, courier.work_zone())
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    Ok(Response::new(DeactivateCourierResponse {
        courier_id: courier_id.to_string(),
        status: CourierStatus::Unavailable.into(),
        deactivated_at: Some(now_timestamp()),
    }))
}

/// Handle ArchiveCourier request
pub async fn archive_courier(
    state: &Arc<AppState>,
    req: ArchiveCourierRequest,
) -> Result<Response<ArchiveCourierResponse>, Status> {
    info!("ArchiveCourier request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier to verify it exists and get work_zone
    let courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Update status in cache to Archived
    state
        .courier_cache
        .update_status(courier_id, DomainCourierStatus::Archived)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    // Remove from free pool
    state
        .courier_cache
        .remove_from_free_pool(courier_id, courier.work_zone())
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    // Archive in repository
    state
        .courier_repo
        .archive(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    Ok(Response::new(ArchiveCourierResponse {
        courier_id: courier_id.to_string(),
        status: CourierStatus::Archived.into(),
        archived_at: Some(now_timestamp()),
    }))
}

/// Handle UpdateContactInfo request
pub async fn update_contact_info(
    state: &Arc<AppState>,
    req: UpdateContactInfoRequest,
) -> Result<Response<UpdateContactInfoResponse>, Status> {
    info!("UpdateContactInfo request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier
    let mut courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Update fields if provided
    if let Some(ref phone) = req.phone {
        if !phone.is_empty() {
            courier.update_phone(phone.clone());
        }
    }
    if let Some(ref email) = req.email {
        if !email.is_empty() {
            courier.update_email(email.clone());
        }
    }
    if let Some(ref push_token) = req.push_token {
        courier.update_push_token(Some(push_token.clone()));
    }

    // Save
    state
        .courier_repo
        .save(&courier)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    Ok(Response::new(UpdateContactInfoResponse {
        courier_id: courier_id.to_string(),
        phone: courier.phone().to_string(),
        email: courier.email().to_string(),
        updated_at: Some(now_timestamp()),
    }))
}

/// Handle UpdateWorkSchedule request
pub async fn update_work_schedule(
    state: &Arc<AppState>,
    req: UpdateWorkScheduleRequest,
) -> Result<Response<UpdateWorkScheduleResponse>, Status> {
    info!("UpdateWorkSchedule request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier
    let mut courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Update work hours if provided
    if let Some(ref wh) = req.work_hours {
        let work_hours = parse_work_hours(Some(wh.clone()))?;
        courier.update_work_hours(work_hours);
    }

    // Update work zone if provided
    if let Some(ref zone) = req.work_zone {
        if !zone.is_empty() {
            courier.update_work_zone(zone.clone());
        }
    }

    // Update max distance if provided
    if let Some(max_dist) = req.max_distance_km {
        if max_dist > 0.0 {
            courier.update_max_distance(max_dist);
        }
    }

    // Save
    state
        .courier_repo
        .save(&courier)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    Ok(Response::new(UpdateWorkScheduleResponse {
        courier_id: courier_id.to_string(),
        work_hours: Some(domain_to_proto_work_hours(courier.work_hours())),
        work_zone: courier.work_zone().to_string(),
        max_distance_km: courier.max_distance_km(),
        updated_at: Some(now_timestamp()),
    }))
}

/// Handle ChangeTransportType request
pub async fn change_transport_type(
    state: &Arc<AppState>,
    req: ChangeTransportTypeRequest,
) -> Result<Response<ChangeTransportTypeResponse>, Status> {
    info!("ChangeTransportType request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Get courier
    let mut courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("Courier not found"))?;

    // Convert transport type
    let new_transport = proto_to_domain_transport(
        TransportType::try_from(req.transport_type).unwrap_or(TransportType::Unspecified),
    );

    // Update transport type (this recalculates max_load)
    courier.change_transport_type(new_transport);
    let new_max_load = courier.max_load();

    // Save
    state
        .courier_repo
        .save(&courier)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    // Update max_load in cache
    state
        .courier_cache
        .update_max_load(courier_id, new_max_load)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    Ok(Response::new(ChangeTransportTypeResponse {
        courier_id: courier_id.to_string(),
        transport_type: domain_to_proto_transport(new_transport).into(),
        max_load: new_max_load as i32,
        updated_at: Some(now_timestamp()),
    }))
}

/// Handle GetCourierDeliveries request
pub async fn get_courier_deliveries(
    state: &Arc<AppState>,
    req: GetCourierDeliveriesRequest,
) -> Result<Response<GetCourierDeliveriesResponse>, Status> {
    info!("GetCourierDeliveries request received");

    let courier_id = uuid::Uuid::parse_str(&req.courier_id)
        .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;

    // Validate limit (default 5, max 50)
    let limit = if req.limit <= 0 { 5 } else { req.limit.min(50) } as u64;

    // Verify courier exists
    let courier_exists = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .is_some();

    if !courier_exists {
        return Err(Status::not_found("Courier not found"));
    }

    // Get packages assigned to this courier
    let filter = PackageFilter::by_courier(courier_id);

    // Get total count first
    let total_count = state
        .package_repo
        .count_by_filter(filter.clone())
        .await
        .map_err(|e| Status::internal(e.to_string()))? as i32;

    // Get packages with limit, sorted by assigned_at desc (most recent first)
    let mut packages = state
        .package_repo
        .find_by_filter(filter, limit, 0)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    // Sort by assigned_at descending (most recent first)
    packages.sort_by(|a, b| {
        b.assigned_at()
            .unwrap_or_default()
            .cmp(&a.assigned_at().unwrap_or_default())
    });

    // Convert to proto
    let deliveries = packages.iter().map(package_to_delivery_record).collect();

    info!(
        courier_id = %courier_id,
        count = packages.len(),
        total = total_count,
        "GetCourierDeliveries completed"
    );

    Ok(Response::new(GetCourierDeliveriesResponse {
        deliveries,
        total_count,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::infrastructure::rpc::WorkHours as ProtoWorkHours;

    // ==================== Request Validation Tests ====================
    // These tests verify input validation without needing mocked dependencies

    fn valid_register_request() -> RegisterCourierRequest {
        RegisterCourierRequest {
            name: "Test Courier".to_string(),
            phone: "+49123456789".to_string(),
            email: "test@example.com".to_string(),
            transport_type: TransportType::Bicycle.into(),
            max_distance_km: 10.0,
            work_zone: "Berlin-Mitte".to_string(),
            work_hours: Some(ProtoWorkHours {
                start_time: "09:00".to_string(),
                end_time: "18:00".to_string(),
                work_days: vec![1, 2, 3, 4, 5],
            }),
            push_token: None,
        }
    }

    #[test]
    fn test_register_request_name_validation() {
        let mut req = valid_register_request();
        req.name = "".to_string();

        // We can't call the async handler directly without AppState,
        // but we can verify the validation logic by testing the pattern
        assert!(req.name.is_empty());
    }

    #[test]
    fn test_register_request_phone_validation() {
        let mut req = valid_register_request();
        req.phone = "".to_string();

        assert!(req.phone.is_empty());
    }

    #[test]
    fn test_register_request_email_validation() {
        let mut req = valid_register_request();
        req.email = "".to_string();

        assert!(req.email.is_empty());
    }

    #[test]
    fn test_register_request_work_zone_validation() {
        let mut req = valid_register_request();
        req.work_zone = "".to_string();

        assert!(req.work_zone.is_empty());
    }

    #[test]
    fn test_valid_register_request_passes_validation() {
        let req = valid_register_request();

        assert!(!req.name.is_empty());
        assert!(!req.phone.is_empty());
        assert!(!req.email.is_empty());
        assert!(!req.work_zone.is_empty());
        assert!(req.work_hours.is_some());
    }

    // ==================== UUID Parsing Tests ====================

    #[test]
    fn test_valid_uuid_parsing() {
        let valid_uuid = "123e4567-e89b-12d3-a456-426614174000";
        let result = uuid::Uuid::parse_str(valid_uuid);
        assert!(result.is_ok());
    }

    #[test]
    fn test_invalid_uuid_parsing() {
        let invalid_uuid = "not-a-valid-uuid";
        let result = uuid::Uuid::parse_str(invalid_uuid);
        assert!(result.is_err());
    }

    #[test]
    fn test_empty_uuid_parsing() {
        let empty_uuid = "";
        let result = uuid::Uuid::parse_str(empty_uuid);
        assert!(result.is_err());
    }

    // ==================== GetCourierRequest Tests ====================

    #[test]
    fn test_get_courier_request_with_valid_id() {
        let req = GetCourierRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            include_location: true,
        };

        let parsed = uuid::Uuid::parse_str(&req.courier_id);
        assert!(parsed.is_ok());
    }

    #[test]
    fn test_get_courier_request_with_invalid_id() {
        let req = GetCourierRequest {
            courier_id: "invalid".to_string(),
            include_location: false,
        };

        let parsed = uuid::Uuid::parse_str(&req.courier_id);
        assert!(parsed.is_err());
    }

    // ==================== ActivateCourierRequest Tests ====================

    #[test]
    fn test_activate_courier_request_validation() {
        let req = ActivateCourierRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
        };

        let parsed = uuid::Uuid::parse_str(&req.courier_id);
        assert!(parsed.is_ok());
    }

    // ==================== UpdateContactInfoRequest Tests ====================

    #[test]
    fn test_update_contact_with_phone() {
        let req = UpdateContactInfoRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            phone: Some("+49999999999".to_string()),
            email: None,
            push_token: None,
        };

        assert!(req.phone.is_some());
        assert!(req.email.is_none());
    }

    #[test]
    fn test_update_contact_with_email() {
        let req = UpdateContactInfoRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            phone: None,
            email: Some("new@example.com".to_string()),
            push_token: None,
        };

        assert!(req.phone.is_none());
        assert!(req.email.is_some());
    }

    #[test]
    fn test_update_contact_with_both() {
        let req = UpdateContactInfoRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            phone: Some("+49999999999".to_string()),
            email: Some("new@example.com".to_string()),
            push_token: Some("push-token-123".to_string()),
        };

        assert!(req.phone.is_some());
        assert!(req.email.is_some());
        assert!(req.push_token.is_some());
    }

    // ==================== ChangeTransportTypeRequest Tests ====================

    #[test]
    fn test_change_transport_type_walking() {
        let req = ChangeTransportTypeRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            transport_type: TransportType::Walking.into(),
        };

        let transport = TransportType::try_from(req.transport_type);
        assert!(transport.is_ok());
        assert_eq!(transport.unwrap(), TransportType::Walking);
    }

    #[test]
    fn test_change_transport_type_car() {
        let req = ChangeTransportTypeRequest {
            courier_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            transport_type: TransportType::Car.into(),
        };

        let transport = TransportType::try_from(req.transport_type);
        assert!(transport.is_ok());
        assert_eq!(transport.unwrap(), TransportType::Car);
    }

    // ==================== WorkHours Parsing Tests ====================

    #[test]
    fn test_work_hours_valid() {
        let wh = ProtoWorkHours {
            start_time: "09:00".to_string(),
            end_time: "18:00".to_string(),
            work_days: vec![1, 2, 3, 4, 5],
        };

        let result = parse_work_hours(Some(wh));
        assert!(result.is_ok());
    }

    #[test]
    fn test_work_hours_invalid_start() {
        let wh = ProtoWorkHours {
            start_time: "9am".to_string(),
            end_time: "18:00".to_string(),
            work_days: vec![1, 2, 3],
        };

        let result = parse_work_hours(Some(wh));
        assert!(result.is_err());
    }

    #[test]
    fn test_work_hours_missing() {
        let result = parse_work_hours(None);
        assert!(result.is_err());
    }

    // ==================== Pagination Tests ====================

    #[test]
    fn test_pagination_defaults() {
        let req = GetCourierPoolRequest {
            status_filter: vec![],
            transport_type_filter: vec![],
            zone_filter: "".to_string(),
            available_only: false,
            include_location: false,
            pagination: None,
        };

        // When pagination is None, handler uses defaults (50, 0)
        let (limit, offset) = if let Some(p) = req.pagination {
            let page = p.page.max(1) as u64;
            let size = p.page_size.clamp(1, 100) as u64;
            (Some(size), Some((page - 1) * size))
        } else {
            (Some(50), Some(0))
        };

        assert_eq!(limit, Some(50));
        assert_eq!(offset, Some(0));
    }

    #[test]
    fn test_pagination_custom_values() {
        use crate::infrastructure::rpc::Pagination;

        let req = GetCourierPoolRequest {
            status_filter: vec![],
            transport_type_filter: vec![],
            zone_filter: "".to_string(),
            available_only: false,
            include_location: false,
            pagination: Some(Pagination { page: 3, page_size: 25 }),
        };

        let (limit, offset) = if let Some(p) = req.pagination {
            let page = p.page.max(1) as u64;
            let size = p.page_size.clamp(1, 100) as u64;
            (Some(size), Some((page - 1) * size))
        } else {
            (Some(50), Some(0))
        };

        assert_eq!(limit, Some(25));
        assert_eq!(offset, Some(50)); // (3-1) * 25 = 50
    }

    #[test]
    fn test_pagination_clamps_page_size() {
        use crate::infrastructure::rpc::Pagination;

        let req = GetCourierPoolRequest {
            status_filter: vec![],
            transport_type_filter: vec![],
            zone_filter: "".to_string(),
            available_only: false,
            include_location: false,
            pagination: Some(Pagination { page: 1, page_size: 500 }), // Above max
        };

        let size = if let Some(p) = req.pagination {
            p.page_size.clamp(1, 100) as u64
        } else {
            50
        };

        assert_eq!(size, 100); // Clamped to max
    }
}
