//! Create Geofence Handler
//!
//! Handles creating a new geofence.

use std::sync::Arc;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::domain::model::{
    Coordinates, CoordinatesError, Geofence, GeofenceError, GeofenceShape,
};
use crate::domain::ports::{CommandHandlerWithResult, GeofenceRepository, RepositoryError};

use super::command::ShapeInput;
use super::Command;

/// Errors that can occur during geofence creation
#[derive(Debug, Error)]
pub enum CreateGeofenceError {
    #[error("Invalid coordinates: {0}")]
    InvalidCoordinates(#[from] CoordinatesError),

    #[error("Invalid geofence: {0}")]
    InvalidGeofence(#[from] GeofenceError),

    #[error("Repository error: {0}")]
    RepositoryError(#[from] RepositoryError),
}

/// Response from creating a geofence
#[derive(Debug, Clone)]
pub struct Response {
    /// The created geofence ID
    pub geofence_id: Uuid,
    /// Name of the geofence
    pub name: String,
    /// Shape type
    pub shape_type: String,
    /// Creation timestamp
    pub created_at: DateTime<Utc>,
}

/// Create Geofence Handler
pub struct Handler<R>
where
    R: GeofenceRepository,
{
    geofence_repo: Arc<R>,
}

impl<R> Handler<R>
where
    R: GeofenceRepository,
{
    /// Create a new handler
    pub fn new(geofence_repo: Arc<R>) -> Self {
        Self { geofence_repo }
    }

    fn shape_type_name(shape: &GeofenceShape) -> String {
        match shape {
            GeofenceShape::Circle { .. } => "circle".to_string(),
            GeofenceShape::Polygon { .. } => "polygon".to_string(),
            GeofenceShape::Rectangle { .. } => "rectangle".to_string(),
        }
    }
}

#[async_trait]
impl<R> CommandHandlerWithResult<Command, Response> for Handler<R>
where
    R: GeofenceRepository + Send + Sync,
{
    type Error = CreateGeofenceError;

    async fn handle(&self, cmd: Command) -> Result<Response, Self::Error> {
        // 1. Create geofence based on shape type
        let geofence = match cmd.shape {
            ShapeInput::Circle {
                center_latitude,
                center_longitude,
                radius_meters,
            } => {
                let center = Coordinates::new(center_latitude, center_longitude)?;
                Geofence::new_circle(cmd.name, cmd.description, center, radius_meters, cmd.trigger)?
            }
            ShapeInput::Polygon { vertices } => {
                let coords: Result<Vec<Coordinates>, _> = vertices
                    .into_iter()
                    .map(|(lat, lon)| Coordinates::new(lat, lon))
                    .collect();
                let coords = coords?;
                Geofence::new_polygon(cmd.name, cmd.description, coords, cmd.trigger)?
            }
            ShapeInput::Rectangle {
                southwest_latitude,
                southwest_longitude,
                northeast_latitude,
                northeast_longitude,
            } => {
                let southwest = Coordinates::new(southwest_latitude, southwest_longitude)?;
                let northeast = Coordinates::new(northeast_latitude, northeast_longitude)?;
                Geofence::new_rectangle(cmd.name, cmd.description, southwest, northeast, cmd.trigger)?
            }
        };

        let geofence_id = geofence.id();
        let name = geofence.name().to_string();
        let shape_type = Self::shape_type_name(geofence.shape());
        let created_at = geofence.created_at();

        // 2. Save to repository
        self.geofence_repo.save(&geofence).await?;

        Ok(Response {
            geofence_id,
            name,
            shape_type,
            created_at,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::model::GeofenceTrigger;
    use async_trait::async_trait;
    use std::collections::HashMap;
    use std::sync::Mutex;

    // === Mock Repository ===

    struct MockGeofenceRepository {
        geofences: Mutex<HashMap<Uuid, Geofence>>,
    }

    impl MockGeofenceRepository {
        fn new() -> Self {
            Self {
                geofences: Mutex::new(HashMap::new()),
            }
        }
    }

    #[async_trait]
    impl GeofenceRepository for MockGeofenceRepository {
        async fn save(&self, geofence: &Geofence) -> Result<(), RepositoryError> {
            let mut geofences = self.geofences.lock().unwrap();
            geofences.insert(geofence.id(), geofence.clone());
            Ok(())
        }

        async fn find_by_id(&self, id: Uuid) -> Result<Option<Geofence>, RepositoryError> {
            let geofences = self.geofences.lock().unwrap();
            Ok(geofences.get(&id).cloned())
        }

        async fn find_active(&self) -> Result<Vec<Geofence>, RepositoryError> {
            let geofences = self.geofences.lock().unwrap();
            Ok(geofences.values().filter(|g| g.is_active()).cloned().collect())
        }

        async fn find_all(&self, _limit: u64, _offset: u64) -> Result<Vec<Geofence>, RepositoryError> {
            let geofences = self.geofences.lock().unwrap();
            Ok(geofences.values().cloned().collect())
        }

        async fn delete(&self, id: Uuid) -> Result<(), RepositoryError> {
            let mut geofences = self.geofences.lock().unwrap();
            geofences.remove(&id);
            Ok(())
        }
    }

    // === Tests ===

    #[tokio::test]
    async fn test_create_circle_geofence() {
        let repo = Arc::new(MockGeofenceRepository::new());
        let handler = Handler::new(repo.clone());

        let cmd = Command::circle(
            "Berlin Center".to_string(),
            Some("Test zone".to_string()),
            52.52,
            13.405,
            1000.0,
            GeofenceTrigger::OnEnter,
        );

        let result = handler.handle(cmd).await;
        assert!(result.is_ok(), "Expected Ok, got: {:?}", result.err());

        let response = result.unwrap();
        assert_eq!(response.name, "Berlin Center");
        assert_eq!(response.shape_type, "circle");

        // Verify saved
        let saved = repo.find_by_id(response.geofence_id).await.unwrap();
        assert!(saved.is_some());
    }

    #[tokio::test]
    async fn test_create_polygon_geofence() {
        let repo = Arc::new(MockGeofenceRepository::new());
        let handler = Handler::new(repo);

        let vertices = vec![
            (52.50, 13.40),
            (52.54, 13.40),
            (52.52, 13.45),
        ];
        let cmd = Command::polygon(
            "Triangle Zone".to_string(),
            None,
            vertices,
            GeofenceTrigger::OnBoth,
        );

        let result = handler.handle(cmd).await;
        assert!(result.is_ok());
        assert_eq!(result.unwrap().shape_type, "polygon");
    }

    #[tokio::test]
    async fn test_create_rectangle_geofence() {
        let repo = Arc::new(MockGeofenceRepository::new());
        let handler = Handler::new(repo);

        let cmd = Command::rectangle(
            "Berlin Area".to_string(),
            Some("Rectangular zone".to_string()),
            52.50,
            13.40,
            52.54,
            13.45,
            GeofenceTrigger::OnExit,
        );

        let result = handler.handle(cmd).await;
        assert!(result.is_ok());
        assert_eq!(result.unwrap().shape_type, "rectangle");
    }

    #[tokio::test]
    async fn test_invalid_coordinates() {
        let repo = Arc::new(MockGeofenceRepository::new());
        let handler = Handler::new(repo);

        let cmd = Command::circle(
            "Invalid".to_string(),
            None,
            91.0, // Invalid latitude
            13.405,
            1000.0,
            GeofenceTrigger::OnEnter,
        );

        let result = handler.handle(cmd).await;
        assert!(matches!(result, Err(CreateGeofenceError::InvalidCoordinates(_))));
    }
}
