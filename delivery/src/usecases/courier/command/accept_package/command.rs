//! Accept Package Command
//!
//! Data structure representing a courier load increment.

use uuid::Uuid;

/// Command to accept a newly assigned package.
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID that should accept the package.
    pub courier_id: Uuid,
}

impl Command {
    /// Create a new command instance.
    pub fn new(courier_id: Uuid) -> Self {
        Self { courier_id }
    }
}
