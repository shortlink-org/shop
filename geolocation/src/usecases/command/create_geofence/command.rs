//! Create Geofence Command
//!
//! Command to create a new geofence.

use crate::domain::model::GeofenceTrigger;

/// Shape type for the geofence
#[derive(Debug, Clone)]
pub enum ShapeInput {
    /// Circle defined by center coordinates and radius in meters
    Circle {
        center_latitude: f64,
        center_longitude: f64,
        radius_meters: f64,
    },
    /// Polygon defined by vertices
    Polygon {
        vertices: Vec<(f64, f64)>, // (latitude, longitude) pairs
    },
    /// Rectangle defined by southwest and northeast corners
    Rectangle {
        southwest_latitude: f64,
        southwest_longitude: f64,
        northeast_latitude: f64,
        northeast_longitude: f64,
    },
}

/// Command to create a geofence
#[derive(Debug, Clone)]
pub struct Command {
    /// Name of the geofence
    pub name: String,
    /// Optional description
    pub description: Option<String>,
    /// Shape of the geofence
    pub shape: ShapeInput,
    /// Trigger type
    pub trigger: GeofenceTrigger,
}

impl Command {
    /// Create a circle geofence command
    pub fn circle(
        name: String,
        description: Option<String>,
        center_latitude: f64,
        center_longitude: f64,
        radius_meters: f64,
        trigger: GeofenceTrigger,
    ) -> Self {
        Self {
            name,
            description,
            shape: ShapeInput::Circle {
                center_latitude,
                center_longitude,
                radius_meters,
            },
            trigger,
        }
    }

    /// Create a polygon geofence command
    pub fn polygon(
        name: String,
        description: Option<String>,
        vertices: Vec<(f64, f64)>,
        trigger: GeofenceTrigger,
    ) -> Self {
        Self {
            name,
            description,
            shape: ShapeInput::Polygon { vertices },
            trigger,
        }
    }

    /// Create a rectangle geofence command
    pub fn rectangle(
        name: String,
        description: Option<String>,
        southwest_latitude: f64,
        southwest_longitude: f64,
        northeast_latitude: f64,
        northeast_longitude: f64,
        trigger: GeofenceTrigger,
    ) -> Self {
        Self {
            name,
            description,
            shape: ShapeInput::Rectangle {
                southwest_latitude,
                southwest_longitude,
                northeast_latitude,
                northeast_longitude,
            },
            trigger,
        }
    }
}
