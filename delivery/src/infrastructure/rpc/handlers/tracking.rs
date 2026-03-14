//! Customer-facing delivery tracking gRPC handlers.

use std::sync::Arc;

use tonic::{Response, Status};
use tracing::info;
use uuid::Uuid;

use crate::di::AppState;
use crate::domain::model::package::PackageStatus;
use crate::domain::ports::{
    CourierCache, CourierRepository, GeolocationService, PackageRepository,
};
use crate::infrastructure::rpc::converters::{
    datetime_to_timestamp, domain_to_proto_package_status, domain_to_proto_status,
    domain_to_proto_transport,
};
use crate::infrastructure::rpc::{
    GetOrderTrackingRequest, GetOrderTrackingResponse, TrackingCourier,
};

/// Handle GetOrderTracking request.
pub async fn get_order_tracking(
    state: &Arc<AppState>,
    req: GetOrderTrackingRequest,
    customer_id: Uuid,
) -> Result<Response<GetOrderTrackingResponse>, Status> {
    info!("GetOrderTracking request received");

    let order_id = Uuid::parse_str(&req.order_id)
        .map_err(|_| Status::invalid_argument("invalid order_id format"))?;

    let response = build_order_tracking_response(state, order_id, customer_id).await?;

    Ok(Response::new(response))
}

pub async fn build_order_tracking_response(
    state: &Arc<AppState>,
    order_id: Uuid,
    customer_id: Uuid,
) -> Result<GetOrderTrackingResponse, Status> {
    let package = state
        .package_repo
        .find_by_order_id(order_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?
        .ok_or_else(|| Status::not_found("tracking not found for order"))?;

    if package.customer_id() != customer_id {
        return Err(Status::not_found("tracking not found for order"));
    }

    let status = package.status();
    let assigned_at = package.assigned_at().map(datetime_to_timestamp);
    let delivered_at = package.delivered_at().map(datetime_to_timestamp);

    let mut response = GetOrderTrackingResponse {
        order_id: package.order_id().to_string(),
        package_id: package.id().0.to_string(),
        status: domain_to_proto_package_status(status).into(),
        courier: None,
        estimated_minutes_remaining: None,
        distance_km_remaining: None,
        estimated_arrival_at: None,
        assigned_at,
        delivered_at,
    };

    let Some(courier_id) = package.courier_id() else {
        return Ok(response);
    };

    let courier = state
        .courier_repo
        .find_by_id(courier_id)
        .await
        .map_err(|e| Status::internal(e.to_string()))?;

    let cached_state = state
        .courier_cache
        .get_state(courier_id)
        .await
        .ok()
        .flatten();
    let current_location = state
        .geolocation_service
        .get_location(courier_id)
        .await
        .ok()
        .flatten();

    let Some(courier) = courier else {
        return Ok(response);
    };

    let courier_status = cached_state
        .as_ref()
        .map(|state| domain_to_proto_status(state.status))
        .unwrap_or_else(|| domain_to_proto_status(courier.status()));

    let mut tracking_courier = TrackingCourier {
        courier_id: courier.id().0.to_string(),
        name: courier.name().to_string(),
        phone: courier.phone().to_string(),
        transport_type: domain_to_proto_transport(courier.transport_type()).into(),
        status: courier_status.into(),
        current_location: current_location.as_ref().map(|location| {
            crate::infrastructure::rpc::Location {
                latitude: location.latitude(),
                longitude: location.longitude(),
            }
        }),
        last_active_at: current_location
            .as_ref()
            .map(|location| datetime_to_timestamp(location.timestamp())),
    };

    if matches!(status, PackageStatus::Assigned | PackageStatus::InTransit) {
        if let Some(location) = current_location.as_ref() {
            let distance_km = location
                .location()
                .distance_to(&package.delivery_address().location);
            let eta_minutes = courier
                .transport_type()
                .calculate_travel_time_minutes(distance_km);
            let eta_minutes_rounded = eta_minutes.ceil() as i32;

            response.distance_km_remaining = Some(distance_km);
            response.estimated_minutes_remaining = Some(eta_minutes_rounded);
            response.estimated_arrival_at = Some(datetime_to_timestamp(
                chrono::Utc::now() + chrono::Duration::minutes(eta_minutes_rounded as i64),
            ));
        }
    }

    if delivered_at.is_some() {
        tracking_courier.current_location = None;
    }

    response.courier = Some(tracking_courier);

    Ok(response)
}
