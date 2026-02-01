//! Activate Courier Command
//!
//! Data structure representing the command to activate a courier.

use uuid::Uuid;

/// Command to activate a courier (set status to FREE)
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to activate
    pub courier_id: Uuid,
}

impl Command {
    /// Create a new ActivateCourier command
    pub fn new(courier_id: Uuid) -> Self {
        Self { courier_id }
    }
}
