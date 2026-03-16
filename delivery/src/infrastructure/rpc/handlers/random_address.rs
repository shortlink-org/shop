//! GetRandomAddress gRPC handler.
//!
//! Samples a random point inside the configured bbox and snaps it to the nearest
//! routable road point through OSRM.

use std::sync::Arc;

use rand::Rng;
use tonic::{Response, Status};
use tracing::instrument;

use crate::di::AppState;
use crate::infrastructure::rpc::{Address, GetRandomAddressRequest, GetRandomAddressResponse};

#[instrument(skip(state))]
pub async fn get_random_address(
    state: &Arc<AppState>,
    _req: GetRandomAddressRequest,
) -> Result<Response<GetRandomAddressResponse>, Status> {
    let bbox = state.random_address_bbox.as_ref().ok_or_else(|| {
        Status::failed_precondition("Random address is not configured (set RANDOM_ADDRESS_* env)")
    })?;

    let (latitude, longitude) = {
        let mut rng = rand::thread_rng();
        (
            rng.gen_range(bbox.min_lat..=bbox.max_lat),
            rng.gen_range(bbox.min_lon..=bbox.max_lon),
        )
    };

    let snapped_point = state
        .osrm_client
        .nearest_driving(latitude, longitude)
        .await
        .map_err(|error| Status::unavailable(format!("OSRM nearest failed: {error}")))?;

    let address = Address {
        street: snapped_point.street.unwrap_or_default(),
        city: bbox.default_city.clone(),
        country: bbox.default_country.clone(),
        latitude: snapped_point.latitude,
        longitude: snapped_point.longitude,
    };

    Ok(Response::new(GetRandomAddressResponse {
        address: Some(address),
    }))
}
