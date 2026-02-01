//! Get Courier Locations Query
//!
//! Query to retrieve current locations for one or more couriers.

use uuid::Uuid;

/// Maximum number of courier IDs per request
pub const MAX_COURIER_IDS: usize = 100;

/// Query to get courier locations
#[derive(Debug, Clone)]
pub struct Query {
    /// List of courier IDs to fetch locations for
    pub courier_ids: Vec<Uuid>,
}

impl Query {
    /// Create a new query
    pub fn new(courier_ids: Vec<Uuid>) -> Self {
        Self { courier_ids }
    }

    /// Create a query for a single courier
    pub fn single(courier_id: Uuid) -> Self {
        Self {
            courier_ids: vec![courier_id],
        }
    }

    /// Validate the query
    pub fn validate(&self) -> Result<(), String> {
        if self.courier_ids.is_empty() {
            return Err("At least one courier ID is required".to_string());
        }
        if self.courier_ids.len() > MAX_COURIER_IDS {
            return Err(format!(
                "Too many courier IDs: {} (max: {})",
                self.courier_ids.len(),
                MAX_COURIER_IDS
            ));
        }
        Ok(())
    }
}
