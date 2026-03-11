//! GetRandomAddress gRPC handler
//!
//! Returns a random address within the configured bounding box (e.g. Berlin).

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

    let mut rng = rand::thread_rng();
    let latitude = rng.gen_range(bbox.min_lat..=bbox.max_lat);
    let longitude = rng.gen_range(bbox.min_lon..=bbox.max_lon);

    let street = format!("Random point ({:.5}, {:.5})", latitude, longitude);

    let address = Address {
        street,
        city: bbox.default_city.clone(),
        postal_code: String::new(),
        country: bbox.default_country.clone(),
        latitude,
        longitude,
    };

    Ok(Response::new(GetRandomAddressResponse {
        address: Some(address),
    }))
}
