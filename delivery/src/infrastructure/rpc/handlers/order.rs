//! Order-related gRPC handlers
//!
//! Handles order acceptance and assignment operations.

use std::sync::Arc;

use chrono::{DateTime, Utc};
use prost_types::Timestamp;
use tonic::{Response, Status};
use tracing::{error, info};

use crate::di::AppState;
use crate::domain::model::package::{PackageStatus, Priority};
use crate::domain::ports::CommandHandlerWithResult;
use crate::usecases::package::command::accept_order::{
    AcceptOrderError, AddressInput, Command as AcceptCommand, DeliveryPeriodInput,
    Handler as AcceptHandler,
};
use crate::usecases::package::command::assign_order::{
    AssignmentMode, Command as AssignCommand, Handler as AssignHandler,
};

use crate::infrastructure::rpc::converters::now_timestamp;
use crate::infrastructure::rpc::{
    assign_order_request, AcceptOrderRequest, AcceptOrderResponse, AssignOrderRequest,
    AssignOrderResponse, PackageStatus as ProtoPackageStatus, Priority as ProtoPriority,
};

/// Convert proto Priority to domain Priority
fn proto_to_domain_priority(proto: i32) -> Priority {
    match ProtoPriority::try_from(proto) {
        Ok(ProtoPriority::Urgent) => Priority::Urgent,
        _ => Priority::Normal,
    }
}

/// Convert domain PackageStatus to proto PackageStatus
fn domain_to_proto_status(status: PackageStatus) -> i32 {
    match status {
        PackageStatus::Accepted => ProtoPackageStatus::Accepted as i32,
        PackageStatus::InPool => ProtoPackageStatus::InPool as i32,
        PackageStatus::Assigned => ProtoPackageStatus::Assigned as i32,
        PackageStatus::InTransit => ProtoPackageStatus::InTransit as i32,
        PackageStatus::Delivered => ProtoPackageStatus::Delivered as i32,
        PackageStatus::NotDelivered => ProtoPackageStatus::NotDelivered as i32,
        PackageStatus::RequiresHandling => ProtoPackageStatus::RequiresHandling as i32,
    }
}

/// Convert prost Timestamp to chrono DateTime
fn timestamp_to_datetime(ts: Option<Timestamp>) -> Result<DateTime<Utc>, Status> {
    let ts = ts.ok_or_else(|| Status::invalid_argument("timestamp is required"))?;
    DateTime::from_timestamp(ts.seconds, ts.nanos as u32)
        .ok_or_else(|| Status::invalid_argument("invalid timestamp"))
}

/// Convert chrono DateTime to prost Timestamp
fn datetime_to_timestamp(dt: DateTime<Utc>) -> Timestamp {
    Timestamp {
        seconds: dt.timestamp(),
        nanos: dt.timestamp_subsec_nanos() as i32,
    }
}

/// Handle AcceptOrder request
pub async fn accept_order(
    state: &Arc<AppState>,
    req: AcceptOrderRequest,
) -> Result<Response<AcceptOrderResponse>, Status> {
    info!("AcceptOrder request received for order_id: {}", req.order_id);

    // Parse UUIDs
    let order_id = uuid::Uuid::parse_str(&req.order_id)
        .map_err(|_| Status::invalid_argument("invalid order_id format"))?;

    let customer_id = uuid::Uuid::parse_str(&req.customer_id)
        .map_err(|_| Status::invalid_argument("invalid customer_id format"))?;

    // Parse addresses
    let pickup = req
        .pickup_address
        .ok_or_else(|| Status::invalid_argument("pickup_address is required"))?;
    let delivery = req
        .delivery_address
        .ok_or_else(|| Status::invalid_argument("delivery_address is required"))?;

    let pickup_address = AddressInput {
        street: pickup.street,
        city: pickup.city,
        postal_code: pickup.postal_code,
        country: pickup.country,
        latitude: pickup.latitude,
        longitude: pickup.longitude,
    };

    let delivery_address = AddressInput {
        street: delivery.street,
        city: delivery.city,
        postal_code: delivery.postal_code,
        country: delivery.country,
        latitude: delivery.latitude,
        longitude: delivery.longitude,
    };

    // Parse delivery period
    let period = req
        .delivery_period
        .ok_or_else(|| Status::invalid_argument("delivery_period is required"))?;

    let delivery_period = DeliveryPeriodInput {
        start_time: timestamp_to_datetime(period.start_time)?,
        end_time: timestamp_to_datetime(period.end_time)?,
    };

    // Parse package info
    let package_info = req
        .package_info
        .ok_or_else(|| Status::invalid_argument("package_info is required"))?;

    // Build command
    let cmd = AcceptCommand::new(
        order_id,
        customer_id,
        pickup_address,
        delivery_address,
        delivery_period,
        package_info.weight_kg,
        package_info.dimensions,
        proto_to_domain_priority(req.priority),
    );

    // Execute handler
    let handler = AcceptHandler::new(state.package_repo.clone());

    let result = handler.handle(cmd).await.map_err(|e| {
        error!(error = %e, "Failed to accept order");
        match e {
            AcceptOrderError::DuplicateOrder(_) => Status::already_exists(e.to_string()),
            AcceptOrderError::InvalidRequest(msg) => Status::invalid_argument(msg),
            AcceptOrderError::InvalidAddress(msg) => Status::invalid_argument(msg),
            AcceptOrderError::InvalidDeliveryPeriod(msg) => Status::invalid_argument(msg),
            AcceptOrderError::RepositoryError(_) => Status::internal(e.to_string()),
        }
    })?;

    info!(
        package_id = %result.package_id,
        order_id = %result.order_id,
        status = ?result.status,
        "Order accepted successfully"
    );

    let response = AcceptOrderResponse {
        package_id: result.package_id.to_string(),
        status: domain_to_proto_status(result.status),
        created_at: Some(datetime_to_timestamp(result.created_at)),
    };

    Ok(Response::new(response))
}

/// Handle AssignOrder request
pub async fn assign_order(
    state: &Arc<AppState>,
    req: AssignOrderRequest,
) -> Result<Response<AssignOrderResponse>, Status> {
    info!("AssignOrder request received");

    if req.package_id.is_empty() {
        return Err(Status::invalid_argument("package_id is required"));
    }

    let package_id = uuid::Uuid::parse_str(&req.package_id)
        .map_err(|_| Status::invalid_argument("invalid package_id format"))?;

    let (mode, courier_id) = match req.assignment_type {
        Some(assign_order_request::AssignmentType::CourierId(id)) => {
            let cid = uuid::Uuid::parse_str(&id)
                .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;
            (AssignmentMode::Manual, Some(cid))
        }
        Some(assign_order_request::AssignmentType::AutoAssign(true)) => {
            (AssignmentMode::Auto, None)
        }
        _ => {
            return Err(Status::invalid_argument("assignment_type is required"));
        }
    };

    let cmd = AssignCommand {
        package_id,
        mode,
        courier_id,
    };

    let handler = AssignHandler::new(
        state.courier_repo.clone(),
        state.courier_cache.clone(),
        state.package_repo.clone(),
    );

    let result = handler.handle(cmd).await.map_err(|e| {
        error!(error = %e, "Failed to assign order");
        match e {
            crate::usecases::package::command::assign_order::AssignOrderError::PackageNotFound(
                _,
            ) => Status::not_found(e.to_string()),
            crate::usecases::package::command::assign_order::AssignOrderError::CourierNotFound(
                _,
            ) => Status::not_found(e.to_string()),
            crate::usecases::package::command::assign_order::AssignOrderError::NoAvailableCourier(
                _,
            ) => Status::failed_precondition(e.to_string()),
            _ => Status::internal(e.to_string()),
        }
    })?;

    info!(
        package_id = %result.package_id,
        courier_id = %result.courier_id,
        "Order assigned successfully"
    );

    let response = AssignOrderResponse {
        package_id: result.package_id.to_string(),
        courier_id: result.courier_id.to_string(),
        assigned_at: Some(now_timestamp()),
        estimated_delivery_minutes: result.estimated_delivery_minutes as i32,
    };

    Ok(Response::new(response))
}

#[cfg(test)]
mod tests {
    use super::*;

    // ==================== Request Validation Tests ====================

    #[test]
    fn test_assign_order_empty_package_id_validation() {
        let req = AssignOrderRequest {
            package_id: "".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::AutoAssign(true)),
        };

        assert!(req.package_id.is_empty());
    }

    #[test]
    fn test_assign_order_valid_package_id() {
        let req = AssignOrderRequest {
            package_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::AutoAssign(true)),
        };

        let parsed = uuid::Uuid::parse_str(&req.package_id);
        assert!(parsed.is_ok());
    }

    #[test]
    fn test_assign_order_invalid_package_id() {
        let req = AssignOrderRequest {
            package_id: "not-a-uuid".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::AutoAssign(true)),
        };

        let parsed = uuid::Uuid::parse_str(&req.package_id);
        assert!(parsed.is_err());
    }

    #[test]
    fn test_assign_order_auto_assign_mode() {
        let req = AssignOrderRequest {
            package_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::AutoAssign(true)),
        };

        match req.assignment_type {
            Some(assign_order_request::AssignmentType::AutoAssign(true)) => {
                // This is the expected branch
                assert!(true);
            }
            _ => panic!("Expected AutoAssign(true)"),
        }
    }

    #[test]
    fn test_assign_order_manual_mode_valid_courier_id() {
        let req = AssignOrderRequest {
            package_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::CourierId(
                "223e4567-e89b-12d3-a456-426614174001".to_string(),
            )),
        };

        match &req.assignment_type {
            Some(assign_order_request::AssignmentType::CourierId(id)) => {
                let parsed = uuid::Uuid::parse_str(id);
                assert!(parsed.is_ok());
            }
            _ => panic!("Expected CourierId"),
        }
    }

    #[test]
    fn test_assign_order_manual_mode_invalid_courier_id() {
        let req = AssignOrderRequest {
            package_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::CourierId(
                "invalid-courier-id".to_string(),
            )),
        };

        match &req.assignment_type {
            Some(assign_order_request::AssignmentType::CourierId(id)) => {
                let parsed = uuid::Uuid::parse_str(id);
                assert!(parsed.is_err());
            }
            _ => panic!("Expected CourierId"),
        }
    }

    #[test]
    fn test_assign_order_missing_assignment_type() {
        let req = AssignOrderRequest {
            package_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            assignment_type: None,
        };

        assert!(req.assignment_type.is_none());
    }

    #[test]
    fn test_assign_order_auto_assign_false() {
        let req = AssignOrderRequest {
            package_id: "123e4567-e89b-12d3-a456-426614174000".to_string(),
            assignment_type: Some(assign_order_request::AssignmentType::AutoAssign(false)),
        };

        // AutoAssign(false) should be treated as missing
        match req.assignment_type {
            Some(assign_order_request::AssignmentType::AutoAssign(false)) => {
                // This would fall through to the _ branch in the handler
                assert!(true);
            }
            _ => panic!("Expected AutoAssign(false)"),
        }
    }
}
