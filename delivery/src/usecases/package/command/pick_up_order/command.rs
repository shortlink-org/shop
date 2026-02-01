//! Pick Up Order Command
//!
//! Data structure representing the command to confirm package pickup by courier.

use uuid::Uuid;

use crate::domain::model::vo::location::Location;

/// Command to confirm package pickup
#[derive(Debug, Clone)]
pub struct Command {
    /// The package ID
    pub package_id: Uuid,
    /// The courier ID
    pub courier_id: Uuid,
    /// Location where the package was picked up
    pub pickup_location: Location,
}

impl Command {
    /// Create a new PickUpOrder command
    pub fn new(package_id: Uuid, courier_id: Uuid, pickup_location: Location) -> Self {
        Self {
            package_id,
            courier_id,
            pickup_location,
        }
    }
}
