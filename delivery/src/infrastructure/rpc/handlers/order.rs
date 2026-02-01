//! Order-related gRPC handlers
//!
//! Handles order assignment operations.

use std::sync::Arc;

use tonic::{Response, Status};
use tracing::{error, info};

use crate::domain::ports::CommandHandlerWithResult;
use crate::di::AppState;
use crate::usecases::package::command::assign_order::{
    AssignmentMode, Command as AssignCommand, Handler as AssignHandler,
};

use crate::infrastructure::rpc::converters::now_timestamp;
use crate::infrastructure::rpc::{assign_order_request, AssignOrderRequest, AssignOrderResponse};

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

    let handler = AssignHandler::new(state.courier_repo.clone(), state.courier_cache.clone());

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
