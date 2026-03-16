//! Complete Courier Delivery Command
//!
//! Data structure representing delivery completion for a courier aggregate.

use uuid::Uuid;

/// Command to complete one courier delivery.
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to update.
    pub courier_id: Uuid,
    /// Whether the delivery succeeded.
    pub success: bool,
}

impl Command {
    /// Create a new command instance.
    pub fn new(courier_id: Uuid, success: bool) -> Self {
        Self {
            courier_id,
            success,
        }
    }
}
