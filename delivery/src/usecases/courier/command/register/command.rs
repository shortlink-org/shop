//! Register Courier Command
//!
//! Data structure representing the command to register a new courier.

use crate::domain::model::courier::WorkHours;
use crate::domain::model::vo::TransportType;

/// Command to register a new courier
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier's full name
    pub name: String,
    /// Courier's phone number
    pub phone: String,
    /// Courier's email address
    pub email: String,
    /// Type of transport the courier uses
    pub transport_type: TransportType,
    /// Maximum distance the courier is willing to travel (km)
    pub max_distance_km: f64,
    /// Work zone identifier
    pub work_zone: String,
    /// Courier's working hours
    pub work_hours: WorkHours,
    /// Optional push notification token
    pub push_token: Option<String>,
}

impl Command {
    /// Create a new RegisterCourier command
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        name: String,
        phone: String,
        email: String,
        transport_type: TransportType,
        max_distance_km: f64,
        work_zone: String,
        work_hours: WorkHours,
        push_token: Option<String>,
    ) -> Self {
        Self {
            name,
            phone,
            email,
            transport_type,
            max_distance_km,
            work_zone,
            work_hours,
            push_token,
        }
    }
}
