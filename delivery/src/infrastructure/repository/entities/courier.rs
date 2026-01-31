//! Courier Entity for Sea-ORM
//!
//! Maps to the `delivery.couriers` table in PostgreSQL.

use chrono::{DateTime, NaiveTime, Utc};
use rust_decimal::Decimal;
use sea_orm::entity::prelude::*;
use serde::{Deserialize, Serialize};

use crate::domain::model::courier::{
    Courier, CourierCapacity, CourierId, CourierStatus, WorkHours,
};
use crate::domain::model::vo::TransportType;

/// Courier database model
#[derive(Clone, Debug, PartialEq, Eq, DeriveEntityModel, Serialize, Deserialize)]
#[sea_orm(table_name = "couriers", schema_name = "delivery")]
pub struct Model {
    #[sea_orm(primary_key, auto_increment = false)]
    pub id: Uuid,
    pub name: String,
    #[sea_orm(unique)]
    pub phone: String,
    #[sea_orm(unique)]
    pub email: String,
    pub transport_type: String,
    pub max_distance_km: Decimal,
    pub work_zone: String,
    pub work_hours_start: NaiveTime,
    pub work_hours_end: NaiveTime,
    pub work_days: Vec<i32>,
    pub push_token: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub version: i32,
}

/// Relations for the Courier entity
#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {}

impl ActiveModelBehavior for ActiveModel {}

// === Conversion from Domain to DB Model ===

impl From<&Courier> for ActiveModel {
    fn from(courier: &Courier) -> Self {
        use sea_orm::ActiveValue::Set;

        ActiveModel {
            id: Set(courier.id().0),
            name: Set(courier.name().to_string()),
            phone: Set(courier.phone().to_string()),
            email: Set(courier.email().to_string()),
            transport_type: Set(transport_type_to_string(courier.transport_type())),
            max_distance_km: Set(Decimal::try_from(courier.max_distance_km()).unwrap_or_default()),
            work_zone: Set(courier.work_zone().to_string()),
            work_hours_start: Set(courier.work_hours().start),
            work_hours_end: Set(courier.work_hours().end),
            work_days: Set(courier.work_hours().days.iter().map(|&d| d as i32).collect()),
            push_token: Set(courier.push_token().map(String::from)),
            created_at: Set(courier.created_at()),
            updated_at: Set(courier.updated_at()),
            version: Set(courier.version() as i32),
        }
    }
}

// === Conversion from DB Model to Domain ===

impl TryFrom<Model> for Courier {
    type Error = String;

    fn try_from(model: Model) -> Result<Self, Self::Error> {
        let transport_type = string_to_transport_type(&model.transport_type)?;

        let work_days: Vec<u8> = model.work_days.iter().map(|&d| d as u8).collect();

        let work_hours = WorkHours::new(model.work_hours_start, model.work_hours_end, work_days)
            .map_err(|e| e.to_string())?;

        // Determine max_load based on transport type
        let max_load = match transport_type {
            TransportType::Walking => 1,
            TransportType::Bicycle => 2,
            TransportType::Motorcycle => 3,
            TransportType::Car => 5,
        };

        // Reconstitute the domain entity
        // Note: status, current_load, rating, etc. are stored in Redis cache
        // Here we use default values; the actual values should be loaded from cache
        Ok(Courier::reconstitute(
            CourierId::from_uuid(model.id),
            model.name,
            model.phone,
            model.email,
            transport_type,
            model
                .max_distance_km
                .try_into()
                .unwrap_or(transport_type.max_recommended_distance_km()),
            model.work_zone,
            work_hours,
            model.push_token,
            CourierStatus::Unavailable, // Default, load from cache
            CourierCapacity::new(max_load),
            0.0, // Rating from cache
            0,   // Successful deliveries from cache
            0,   // Failed deliveries from cache
            model.created_at,
            model.updated_at,
            model.version as u32,
        ))
    }
}

// === Helper Functions ===

fn transport_type_to_string(tt: TransportType) -> String {
    match tt {
        TransportType::Walking => "walking".to_string(),
        TransportType::Bicycle => "bicycle".to_string(),
        TransportType::Motorcycle => "motorcycle".to_string(),
        TransportType::Car => "car".to_string(),
    }
}

fn string_to_transport_type(s: &str) -> Result<TransportType, String> {
    match s.to_lowercase().as_str() {
        "walking" => Ok(TransportType::Walking),
        "bicycle" => Ok(TransportType::Bicycle),
        "motorcycle" => Ok(TransportType::Motorcycle),
        "car" => Ok(TransportType::Car),
        _ => Err(format!("Unknown transport type: {}", s)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_transport_type_conversion() {
        assert_eq!(
            transport_type_to_string(TransportType::Walking),
            "walking"
        );
        assert_eq!(
            transport_type_to_string(TransportType::Bicycle),
            "bicycle"
        );
        assert_eq!(
            transport_type_to_string(TransportType::Motorcycle),
            "motorcycle"
        );
        assert_eq!(transport_type_to_string(TransportType::Car), "car");

        assert_eq!(
            string_to_transport_type("walking").unwrap(),
            TransportType::Walking
        );
        assert_eq!(
            string_to_transport_type("BICYCLE").unwrap(),
            TransportType::Bicycle
        );
        assert!(string_to_transport_type("unknown").is_err());
    }
}
