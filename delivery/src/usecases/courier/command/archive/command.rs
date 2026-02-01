//! Archive Courier Command
//!
//! Data structure representing the command to archive a courier.

use uuid::Uuid;

/// Command to archive a courier (soft delete)
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to archive
    pub courier_id: Uuid,
    /// Optional reason for archival
    pub reason: Option<String>,
}

impl Command {
    /// Create a new ArchiveCourier command
    pub fn new(courier_id: Uuid, reason: Option<String>) -> Self {
        Self { courier_id, reason }
    }
}
