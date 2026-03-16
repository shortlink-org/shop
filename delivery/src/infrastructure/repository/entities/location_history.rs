//! Location History Entity
//!
//! Sea-ORM entity for the courier_location_history table.

use chrono::{DateTime, Utc};
use sea_orm::entity::prelude::*;
use uuid::Uuid;

use crate::domain::model::courier_location::LocationHistoryEntry;
use crate::domain::model::vo::location::Location;

/// Location history model
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(schema_name = "delivery", table_name = "courier_location_history")]
pub struct Model {
    #[sea_orm(primary_key, auto_increment = false)]
    pub id: Uuid,
    pub courier_id: Uuid,
    pub latitude: f64,
    pub longitude: f64,
    pub accuracy: f64,
    pub timestamp: DateTime<Utc>,
    pub speed: Option<f64>,
    pub heading: Option<f64>,
    pub created_at: DateTime<Utc>,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(
        belongs_to = "super::courier::Entity",
        from = "Column::CourierId",
        to = "super::courier::Column::Id"
    )]
    Courier,
}

impl Related<super::courier::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::Courier.def()
    }
}

impl ActiveModelBehavior for ActiveModel {}

/// Convert from domain model to active model
impl From<&LocationHistoryEntry> for ActiveModel {
    fn from(entry: &LocationHistoryEntry) -> Self {
        use sea_orm::ActiveValue::Set;

        Self {
            id: Set(entry.entry_id()),
            courier_id: Set(entry.reported_by()),
            latitude: Set(entry.reported_position().latitude()),
            longitude: Set(entry.reported_position().longitude()),
            accuracy: Set(entry.reported_position().accuracy()),
            timestamp: Set(entry.recorded_at()),
            speed: Set(entry.travel_speed_kmh()),
            heading: Set(entry.bearing_degrees()),
            created_at: Set(entry.stored_at()),
        }
    }
}

/// Convert from database model to domain model
impl TryFrom<Model> for LocationHistoryEntry {
    type Error = String;

    fn try_from(model: Model) -> Result<Self, Self::Error> {
        let location = Location::new(model.latitude, model.longitude, model.accuracy)
            .map_err(|e| format!("Invalid location: {}", e))?;

        Ok(LocationHistoryEntry::reconstitute(
            model.id,
            model.courier_id,
            location,
            model.timestamp,
            model.speed,
            model.heading,
            model.created_at,
        ))
    }
}
