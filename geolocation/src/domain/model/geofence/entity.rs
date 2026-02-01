//! Geofence Entity
//!
//! Represents a virtual geographic boundary for tracking courier entry/exit.

use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::vo::Coordinates;

/// Errors for geofence operations
#[derive(Debug, Error, Clone, PartialEq)]
pub enum GeofenceError {
    #[error("Invalid radius: must be > 0")]
    InvalidRadius,
    #[error("Polygon must have at least 3 points")]
    InsufficientPolygonPoints,
    #[error("Rectangle requires exactly 2 points (southwest, northeast)")]
    InvalidRectangle,
}

/// Type of geofence shape
#[derive(Debug, Clone, PartialEq)]
pub enum GeofenceShape {
    /// Circle defined by center and radius in meters
    Circle { center: Coordinates, radius_meters: f64 },
    /// Polygon defined by a list of vertices
    Polygon { vertices: Vec<Coordinates> },
    /// Rectangle defined by southwest and northeast corners
    Rectangle { southwest: Coordinates, northeast: Coordinates },
}

/// Geofence trigger type
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum GeofenceTrigger {
    /// Trigger on entry
    OnEnter,
    /// Trigger on exit
    OnExit,
    /// Trigger on both entry and exit
    OnBoth,
}

/// Geofence aggregate root
#[derive(Debug, Clone)]
pub struct Geofence {
    id: Uuid,
    name: String,
    description: Option<String>,
    shape: GeofenceShape,
    trigger: GeofenceTrigger,
    is_active: bool,
    created_at: DateTime<Utc>,
    updated_at: DateTime<Utc>,
}

impl Geofence {
    /// Create a new circular geofence
    pub fn new_circle(
        name: String,
        description: Option<String>,
        center: Coordinates,
        radius_meters: f64,
        trigger: GeofenceTrigger,
    ) -> Result<Self, GeofenceError> {
        if radius_meters <= 0.0 {
            return Err(GeofenceError::InvalidRadius);
        }

        let now = Utc::now();
        Ok(Self {
            id: Uuid::new_v4(),
            name,
            description,
            shape: GeofenceShape::Circle {
                center,
                radius_meters,
            },
            trigger,
            is_active: true,
            created_at: now,
            updated_at: now,
        })
    }

    /// Create a new polygon geofence
    pub fn new_polygon(
        name: String,
        description: Option<String>,
        vertices: Vec<Coordinates>,
        trigger: GeofenceTrigger,
    ) -> Result<Self, GeofenceError> {
        if vertices.len() < 3 {
            return Err(GeofenceError::InsufficientPolygonPoints);
        }

        let now = Utc::now();
        Ok(Self {
            id: Uuid::new_v4(),
            name,
            description,
            shape: GeofenceShape::Polygon { vertices },
            trigger,
            is_active: true,
            created_at: now,
            updated_at: now,
        })
    }

    /// Create a new rectangle geofence
    pub fn new_rectangle(
        name: String,
        description: Option<String>,
        southwest: Coordinates,
        northeast: Coordinates,
        trigger: GeofenceTrigger,
    ) -> Result<Self, GeofenceError> {
        // Validate that northeast is actually northeast of southwest
        if southwest.latitude() >= northeast.latitude()
            || southwest.longitude() >= northeast.longitude()
        {
            return Err(GeofenceError::InvalidRectangle);
        }

        let now = Utc::now();
        Ok(Self {
            id: Uuid::new_v4(),
            name,
            description,
            shape: GeofenceShape::Rectangle {
                southwest,
                northeast,
            },
            trigger,
            is_active: true,
            created_at: now,
            updated_at: now,
        })
    }

    /// Reconstitute from storage
    #[allow(clippy::too_many_arguments)]
    pub fn reconstitute(
        id: Uuid,
        name: String,
        description: Option<String>,
        shape: GeofenceShape,
        trigger: GeofenceTrigger,
        is_active: bool,
        created_at: DateTime<Utc>,
        updated_at: DateTime<Utc>,
    ) -> Self {
        Self {
            id,
            name,
            description,
            shape,
            trigger,
            is_active,
            created_at,
            updated_at,
        }
    }

    /// Check if a point is inside this geofence
    pub fn contains(&self, point: &Coordinates) -> bool {
        match &self.shape {
            GeofenceShape::Circle { center, radius_meters } => {
                let distance_km = center.distance_to(point);
                let distance_meters = distance_km * 1000.0;
                distance_meters <= *radius_meters
            }
            GeofenceShape::Rectangle { southwest, northeast } => {
                point.latitude() >= southwest.latitude()
                    && point.latitude() <= northeast.latitude()
                    && point.longitude() >= southwest.longitude()
                    && point.longitude() <= northeast.longitude()
            }
            GeofenceShape::Polygon { vertices } => {
                // Ray casting algorithm for point-in-polygon
                Self::point_in_polygon(point, vertices)
            }
        }
    }

    /// Ray casting algorithm for point-in-polygon test
    fn point_in_polygon(point: &Coordinates, vertices: &[Coordinates]) -> bool {
        let mut inside = false;
        let n = vertices.len();

        let mut j = n - 1;
        for i in 0..n {
            let vi = &vertices[i];
            let vj = &vertices[j];

            if ((vi.latitude() > point.latitude()) != (vj.latitude() > point.latitude()))
                && (point.longitude()
                    < (vj.longitude() - vi.longitude()) * (point.latitude() - vi.latitude())
                        / (vj.latitude() - vi.latitude())
                        + vi.longitude())
            {
                inside = !inside;
            }
            j = i;
        }

        inside
    }

    // === State mutations ===

    pub fn activate(&mut self) {
        self.is_active = true;
        self.touch();
    }

    pub fn deactivate(&mut self) {
        self.is_active = false;
        self.touch();
    }

    pub fn update_name(&mut self, name: String) {
        self.name = name;
        self.touch();
    }

    fn touch(&mut self) {
        self.updated_at = Utc::now();
    }

    // === Getters ===

    pub fn id(&self) -> Uuid {
        self.id
    }

    pub fn name(&self) -> &str {
        &self.name
    }

    pub fn description(&self) -> Option<&str> {
        self.description.as_deref()
    }

    pub fn shape(&self) -> &GeofenceShape {
        &self.shape
    }

    pub fn trigger(&self) -> GeofenceTrigger {
        self.trigger
    }

    pub fn is_active(&self) -> bool {
        self.is_active
    }

    pub fn created_at(&self) -> DateTime<Utc> {
        self.created_at
    }

    pub fn updated_at(&self) -> DateTime<Utc> {
        self.updated_at
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_circle_geofence() {
        let center = Coordinates::new(52.52, 13.405).unwrap();
        let geofence = Geofence::new_circle(
            "Berlin Center".to_string(),
            Some("Test zone".to_string()),
            center,
            1000.0, // 1km radius
            GeofenceTrigger::OnEnter,
        );
        assert!(geofence.is_ok());

        let geofence = geofence.unwrap();
        assert_eq!(geofence.name(), "Berlin Center");
        assert!(geofence.is_active());

        // Point inside
        let inside = Coordinates::new(52.521, 13.406).unwrap();
        assert!(geofence.contains(&inside));

        // Point outside
        let outside = Coordinates::new(52.55, 13.5).unwrap();
        assert!(!geofence.contains(&outside));
    }

    #[test]
    fn test_rectangle_geofence() {
        let southwest = Coordinates::new(52.50, 13.40).unwrap();
        let northeast = Coordinates::new(52.54, 13.45).unwrap();
        let geofence = Geofence::new_rectangle(
            "Berlin Area".to_string(),
            None,
            southwest,
            northeast,
            GeofenceTrigger::OnBoth,
        );
        assert!(geofence.is_ok());

        let geofence = geofence.unwrap();

        // Point inside
        let inside = Coordinates::new(52.52, 13.42).unwrap();
        assert!(geofence.contains(&inside));

        // Point outside
        let outside = Coordinates::new(52.60, 13.50).unwrap();
        assert!(!geofence.contains(&outside));
    }

    #[test]
    fn test_polygon_geofence() {
        // Triangle around a point
        let vertices = vec![
            Coordinates::new(52.50, 13.40).unwrap(),
            Coordinates::new(52.54, 13.40).unwrap(),
            Coordinates::new(52.52, 13.45).unwrap(),
        ];
        let geofence = Geofence::new_polygon(
            "Triangle Zone".to_string(),
            None,
            vertices,
            GeofenceTrigger::OnExit,
        );
        assert!(geofence.is_ok());

        let geofence = geofence.unwrap();

        // Point inside triangle
        let inside = Coordinates::new(52.52, 13.42).unwrap();
        assert!(geofence.contains(&inside));
    }

    #[test]
    fn test_invalid_geofences() {
        // Invalid radius
        let center = Coordinates::new(52.52, 13.405).unwrap();
        assert!(matches!(
            Geofence::new_circle("Test".to_string(), None, center, 0.0, GeofenceTrigger::OnEnter),
            Err(GeofenceError::InvalidRadius)
        ));

        // Insufficient polygon points
        let vertices = vec![
            Coordinates::new(52.50, 13.40).unwrap(),
            Coordinates::new(52.54, 13.40).unwrap(),
        ];
        assert!(matches!(
            Geofence::new_polygon("Test".to_string(), None, vertices, GeofenceTrigger::OnEnter),
            Err(GeofenceError::InsufficientPolygonPoints)
        ));
    }

    #[test]
    fn test_geofence_activation() {
        let center = Coordinates::new(52.52, 13.405).unwrap();
        let mut geofence = Geofence::new_circle(
            "Test".to_string(),
            None,
            center,
            100.0,
            GeofenceTrigger::OnEnter,
        )
        .unwrap();

        assert!(geofence.is_active());

        geofence.deactivate();
        assert!(!geofence.is_active());

        geofence.activate();
        assert!(geofence.is_active());
    }
}
