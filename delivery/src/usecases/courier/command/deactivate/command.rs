//! Deactivate Courier Command
//!
//! Data structure representing the command to deactivate a courier.

use uuid::Uuid;

/// Command to deactivate a courier (set status to UNAVAILABLE)
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to deactivate
    pub courier_id: Uuid,
    /// Optional reason for deactivation
    pub reason: Option<String>,
}

impl Command {
    /// Create a new DeactivateCourier command
    pub fn new(courier_id: Uuid, reason: Option<String>) -> Self {
        Self { courier_id, reason }
    }
}
