//! Get Courier Query
//!
//! Data structure representing the query to get a single courier.

use uuid::Uuid;

/// Query to get a single courier by ID
#[derive(Debug, Clone)]
pub struct Query {
    /// Courier ID to retrieve
    pub courier_id: Uuid,
    /// Whether to include current location from Geolocation Service
    pub include_location: bool,
}

impl Query {
    /// Create a new GetCourier query
    pub fn new(courier_id: Uuid, include_location: bool) -> Self {
        Self {
            courier_id,
            include_location,
        }
    }
}
