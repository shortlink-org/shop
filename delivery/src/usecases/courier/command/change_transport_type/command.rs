//! Change Transport Type Command
//!
//! Data structure representing the command to change courier transport type.

use uuid::Uuid;

use crate::domain::model::vo::TransportType;

/// Command to change courier transport type
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to update
    pub courier_id: Uuid,
    /// New transport type
    pub transport_type: TransportType,
}

impl Command {
    /// Create a new ChangeTransportType command
    pub fn new(courier_id: Uuid, transport_type: TransportType) -> Self {
        Self {
            courier_id,
            transport_type,
        }
    }
}
