//! gRPC Server Implementation
//!
//! Implements the DeliveryService gRPC interface.

use std::sync::Arc;

use chrono::NaiveTime;
use tonic::{Request, Response, Status};
use tracing::{error, info, instrument};

use crate::boundary::ports::{CommandHandlerWithResult, QueryHandler};
use crate::di::AppState;
use crate::domain::model::courier::WorkHours as DomainWorkHours;
use crate::domain::model::vo::TransportType as DomainTransportType;
use crate::domain::model::courier::CourierStatus as DomainCourierStatus;
use crate::usecases::courier::command::register::{Command as RegisterCommand, Handler as RegisterHandler};
use crate::usecases::courier::query::get_pool::{Handler as GetPoolHandler, Query as GetPoolQuery, CourierFilter};
use crate::usecases::package::command::assign_order::{Command as AssignCommand, Handler as AssignHandler, AssignmentMode};

use super::{
    AssignOrderRequest, AssignOrderResponse,
    Courier as ProtoCourier, CourierStatus, 
    DeliveryService, GetCourierPoolRequest, GetCourierPoolResponse,
    PaginationInfo, RegisterCourierRequest, RegisterCourierResponse,
    TransportType, WorkHours as ProtoWorkHours,
};

/// gRPC service implementation
pub struct DeliveryServiceImpl {
    state: Arc<AppState>,
}

impl DeliveryServiceImpl {
    /// Create a new service instance
    pub fn new(state: Arc<AppState>) -> Self {
        Self { state }
    }

    /// Convert proto TransportType to domain
    fn proto_to_domain_transport(t: TransportType) -> DomainTransportType {
        match t {
            TransportType::Walking => DomainTransportType::Walking,
            TransportType::Bicycle => DomainTransportType::Bicycle,
            TransportType::Motorcycle => DomainTransportType::Motorcycle,
            TransportType::Car => DomainTransportType::Car,
            TransportType::Unspecified => DomainTransportType::Walking,
        }
    }

    /// Convert domain TransportType to proto
    fn domain_to_proto_transport(t: DomainTransportType) -> TransportType {
        match t {
            DomainTransportType::Walking => TransportType::Walking,
            DomainTransportType::Bicycle => TransportType::Bicycle,
            DomainTransportType::Motorcycle => TransportType::Motorcycle,
            DomainTransportType::Car => TransportType::Car,
        }
    }

    /// Convert domain CourierStatus to proto
    fn domain_to_proto_status(s: DomainCourierStatus) -> CourierStatus {
        match s {
            DomainCourierStatus::Unavailable => CourierStatus::Unavailable,
            DomainCourierStatus::Free => CourierStatus::Free,
            DomainCourierStatus::Busy => CourierStatus::Busy,
        }
    }

    /// Convert proto CourierStatus to domain
    fn proto_to_domain_status(s: CourierStatus) -> Option<DomainCourierStatus> {
        match s {
            CourierStatus::Unavailable => Some(DomainCourierStatus::Unavailable),
            CourierStatus::Free => Some(DomainCourierStatus::Free),
            CourierStatus::Busy => Some(DomainCourierStatus::Busy),
            CourierStatus::Unspecified => None,
        }
    }

    /// Parse work hours from proto
    fn parse_work_hours(wh: Option<ProtoWorkHours>) -> Result<DomainWorkHours, Status> {
        let wh = wh.ok_or_else(|| Status::invalid_argument("work_hours is required"))?;
        
        let start = NaiveTime::parse_from_str(&wh.start_time, "%H:%M")
            .map_err(|_| Status::invalid_argument("invalid start_time format, use HH:MM"))?;
        let end = NaiveTime::parse_from_str(&wh.end_time, "%H:%M")
            .map_err(|_| Status::invalid_argument("invalid end_time format, use HH:MM"))?;
        
        let work_days: Vec<u8> = wh.work_days.iter().map(|&d| d as u8).collect();
        
        DomainWorkHours::new(start, end, work_days)
            .map_err(|e| Status::invalid_argument(format!("invalid work_hours: {}", e)))
    }
}

#[tonic::async_trait]
impl DeliveryService for DeliveryServiceImpl {
    #[instrument(skip(self, request), fields(email = %request.get_ref().email))]
    async fn register_courier(
        &self,
        request: Request<RegisterCourierRequest>,
    ) -> Result<Response<RegisterCourierResponse>, Status> {
        let req = request.into_inner();
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

        let transport_type = Self::proto_to_domain_transport(req.transport_type());
        let work_hours = Self::parse_work_hours(req.work_hours)?;

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

        let handler = RegisterHandler::new(
            self.state.courier_repo.clone(),
            self.state.courier_cache.clone(),
        );

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

    #[instrument(skip(self, request))]
    async fn get_courier_pool(
        &self,
        request: Request<GetCourierPoolRequest>,
    ) -> Result<Response<GetCourierPoolResponse>, Status> {
        let req = request.into_inner();
        info!("GetCourierPool request received");

        // Build filter
        let status = req.status_filter.first()
            .and_then(|&s| Self::proto_to_domain_status(CourierStatus::try_from(s).unwrap_or(CourierStatus::Unspecified)));
        
        let transport_type = req.transport_type_filter.first()
            .map(|&t| Self::proto_to_domain_transport(TransportType::try_from(t).unwrap_or(TransportType::Unspecified)));

        let filter = CourierFilter {
            status,
            transport_type,
            work_zone: if req.zone_filter.is_empty() { None } else { Some(req.zone_filter) },
            available_only: req.available_only,
        };

        let (limit, offset) = if let Some(p) = req.pagination {
            let page = p.page.max(1) as u64;
            let size = p.page_size.clamp(1, 100) as u64;
            (Some(size), Some((page - 1) * size))
        } else {
            (Some(50), Some(0))
        };

        let query = GetPoolQuery { filter, limit, offset };

        let handler = GetPoolHandler::new(
            self.state.courier_repo.clone(),
            self.state.courier_cache.clone(),
        );

        let result = handler.handle(query).await.map_err(|e| {
            error!(error = %e, "Failed to get courier pool");
            Status::internal(e.to_string())
        })?;

        let couriers: Vec<ProtoCourier> = result.couriers.into_iter().map(|cws| {
            let c = cws.courier;
            let state = cws.state;
            
            let wh = c.work_hours();
            let work_hours = Some(ProtoWorkHours {
                start_time: wh.start.format("%H:%M").to_string(),
                end_time: wh.end.format("%H:%M").to_string(),
                work_days: wh.days.iter().map(|&d| d as i32).collect(),
            });

            let created_at = c.created_at();

            ProtoCourier {
                courier_id: c.id().0.to_string(),
                name: c.name().to_string(),
                phone: c.phone().to_string(),
                email: c.email().to_string(),
                transport_type: Self::domain_to_proto_transport(c.transport_type()).into(),
                max_distance_km: c.max_distance_km(),
                status: state.as_ref()
                    .map(|s| Self::domain_to_proto_status(s.status))
                    .unwrap_or(CourierStatus::Unavailable)
                    .into(),
                current_load: state.as_ref().map(|s| s.current_load as i32).unwrap_or(0),
                max_load: c.max_load() as i32,
                rating: state.as_ref().map(|s| s.rating).unwrap_or(0.0),
                work_hours,
                work_zone: c.work_zone().to_string(),
                current_location: None, // TODO: integrate with Geolocation Service
                successful_deliveries: state.as_ref().map(|s| s.successful_deliveries as i32).unwrap_or(0),
                failed_deliveries: state.as_ref().map(|s| s.failed_deliveries as i32).unwrap_or(0),
                created_at: Some(prost_types::Timestamp {
                    seconds: created_at.timestamp(),
                    nanos: created_at.timestamp_subsec_nanos() as i32,
                }),
                last_active_at: None,
            }
        }).collect();

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

    #[instrument(skip(self, request), fields(package_id = %request.get_ref().package_id))]
    async fn assign_order(
        &self,
        request: Request<AssignOrderRequest>,
    ) -> Result<Response<AssignOrderResponse>, Status> {
        let req = request.into_inner();
        info!("AssignOrder request received");

        if req.package_id.is_empty() {
            return Err(Status::invalid_argument("package_id is required"));
        }

        let package_id = uuid::Uuid::parse_str(&req.package_id)
            .map_err(|_| Status::invalid_argument("invalid package_id format"))?;

        let (mode, courier_id) = match req.assignment_type {
            Some(super::assign_order_request::AssignmentType::CourierId(id)) => {
                let cid = uuid::Uuid::parse_str(&id)
                    .map_err(|_| Status::invalid_argument("invalid courier_id format"))?;
                (AssignmentMode::Manual, Some(cid))
            }
            Some(super::assign_order_request::AssignmentType::AutoAssign(true)) => {
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
            self.state.courier_repo.clone(),
            self.state.courier_cache.clone(),
        );

        let result = handler.handle(cmd).await.map_err(|e| {
            error!(error = %e, "Failed to assign order");
            match e {
                crate::usecases::package::command::assign_order::AssignOrderError::PackageNotFound(_) => {
                    Status::not_found(e.to_string())
                }
                crate::usecases::package::command::assign_order::AssignOrderError::CourierNotFound(_) => {
                    Status::not_found(e.to_string())
                }
                crate::usecases::package::command::assign_order::AssignOrderError::NoAvailableCourier(_) => {
                    Status::failed_precondition(e.to_string())
                }
                _ => Status::internal(e.to_string()),
            }
        })?;

        info!(
            package_id = %result.package_id,
            courier_id = %result.courier_id,
            "Order assigned successfully"
        );

        let now = chrono::Utc::now();
        let response = AssignOrderResponse {
            package_id: result.package_id.to_string(),
            courier_id: result.courier_id.to_string(),
            assigned_at: Some(prost_types::Timestamp {
                seconds: now.timestamp(),
                nanos: now.timestamp_subsec_nanos() as i32,
            }),
            estimated_delivery_minutes: result.estimated_delivery_minutes as i32,
        };

        Ok(Response::new(response))
    }
}
