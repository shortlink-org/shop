//! Get Courier Location Query
//!
//! Data structure representing the query to get a courier's current location.

use uuid::Uuid;

/// Query to get a courier's current location
#[derive(Debug, Clone)]
pub struct Query {
    /// Courier ID
    pub courier_id: Uuid,
}

impl Query {
    /// Create a new GetCourierLocation query
    pub fn new(courier_id: Uuid) -> Self {
        Self { courier_id }
    }
}

/// Query to get locations for multiple couriers
#[derive(Debug, Clone)]
pub struct BatchQuery {
    /// Courier IDs to retrieve locations for
    pub courier_ids: Vec<Uuid>,
}

impl BatchQuery {
    /// Create a new batch query
    pub fn new(courier_ids: Vec<Uuid>) -> Self {
        Self { courier_ids }
    }
}
