//! Update Work Schedule Command
//!
//! Data structure representing the command to update courier work schedule.

use uuid::Uuid;

use crate::domain::model::courier::WorkHours;

/// Command to update courier work schedule
#[derive(Debug, Clone)]
pub struct Command {
    /// Courier ID to update
    pub courier_id: Uuid,
    /// New work hours (optional)
    pub work_hours: Option<WorkHours>,
    /// New work zone (optional)
    pub work_zone: Option<String>,
    /// New max distance in km (optional)
    pub max_distance_km: Option<f64>,
}

impl Command {
    /// Create a new UpdateWorkSchedule command
    pub fn new(
        courier_id: Uuid,
        work_hours: Option<WorkHours>,
        work_zone: Option<String>,
        max_distance_km: Option<f64>,
    ) -> Self {
        Self {
            courier_id,
            work_hours,
            work_zone,
            max_distance_km,
        }
    }
}
