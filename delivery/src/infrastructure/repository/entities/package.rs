//! Package Entity for Sea-ORM
//!
//! Database entity for the packages table.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sea_orm::entity::prelude::*;
use uuid::Uuid;

use crate::domain::model::package::{
    Address, DeliveryPeriod, Package, PackageError, PackageId, PackageStatus, Priority,
};
use crate::domain::model::vo::location::Location;

/// Package entity for Sea-ORM
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(schema_name = "delivery", table_name = "packages")]
pub struct Model {
    #[sea_orm(primary_key, auto_increment = false)]
    pub id: Uuid,
    pub order_id: Uuid,
    pub customer_id: Uuid,
    // Pickup address
    pub pickup_street: String,
    pub pickup_city: String,
    pub pickup_postal_code: String,
    pub pickup_latitude: f64,
    pub pickup_longitude: f64,
    // Delivery address
    pub delivery_street: String,
    pub delivery_city: String,
    pub delivery_postal_code: String,
    pub delivery_latitude: f64,
    pub delivery_longitude: f64,
    // Delivery period
    pub delivery_period_start: DateTime<Utc>,
    pub delivery_period_end: DateTime<Utc>,
    // Package info
    pub weight_kg: Decimal,
    pub priority: String,
    pub status: String,
    pub courier_id: Option<Uuid>,
    pub zone: String,
    // Timestamps
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub assigned_at: Option<DateTime<Utc>>,
    pub delivered_at: Option<DateTime<Utc>>,
    pub not_delivered_reason: Option<String>,
    pub version: i32,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {}

impl ActiveModelBehavior for ActiveModel {}

impl From<&Package> for ActiveModel {
    fn from(package: &Package) -> Self {
        use sea_orm::ActiveValue::Set;

        let pickup = package.pickup_address();
        let delivery = package.delivery_address();
        let period = package.delivery_period();

        Self {
            id: Set(package.id().0),
            order_id: Set(package.order_id()),
            customer_id: Set(package.customer_id()),
            pickup_street: Set(pickup.street.clone()),
            pickup_city: Set(pickup.city.clone()),
            pickup_postal_code: Set(pickup.postal_code.clone()),
            pickup_latitude: Set(pickup.location.latitude()),
            pickup_longitude: Set(pickup.location.longitude()),
            delivery_street: Set(delivery.street.clone()),
            delivery_city: Set(delivery.city.clone()),
            delivery_postal_code: Set(delivery.postal_code.clone()),
            delivery_latitude: Set(delivery.location.latitude()),
            delivery_longitude: Set(delivery.location.longitude()),
            delivery_period_start: Set(period.start),
            delivery_period_end: Set(period.end),
            weight_kg: Set(Decimal::from_f64_retain(package.weight_kg()).unwrap_or_default()),
            priority: Set(priority_to_string(package.priority())),
            status: Set(status_to_string(package.status())),
            courier_id: Set(package.courier_id()),
            zone: Set(package.zone().to_string()),
            created_at: Set(package.created_at()),
            updated_at: Set(package.updated_at()),
            assigned_at: Set(package.assigned_at()),
            delivered_at: Set(package.delivered_at()),
            not_delivered_reason: Set(package.not_delivered_reason().map(|s| s.to_string())),
            version: Set(package.version() as i32),
        }
    }
}

impl TryFrom<Model> for Package {
    type Error = String;

    fn try_from(model: Model) -> Result<Self, Self::Error> {
        let pickup_location = Location::new(model.pickup_latitude, model.pickup_longitude, 10.0)
            .map_err(|e| e.to_string())?;
        let delivery_location =
            Location::new(model.delivery_latitude, model.delivery_longitude, 10.0)
                .map_err(|e| e.to_string())?;

        let pickup_address = Address::new(
            model.pickup_street,
            model.pickup_city,
            model.pickup_postal_code,
            pickup_location,
        );

        let delivery_address = Address::new(
            model.delivery_street,
            model.delivery_city,
            model.delivery_postal_code,
            delivery_location,
        );

        let delivery_period =
            DeliveryPeriod::new(model.delivery_period_start, model.delivery_period_end)
                .map_err(|e: PackageError| e.to_string())?;

        let priority = string_to_priority(&model.priority);
        let status = string_to_status(&model.status);

        Ok(Package::reconstitute(
            PackageId::from_uuid(model.id),
            model.order_id,
            model.customer_id,
            pickup_address,
            delivery_address,
            delivery_period,
            model
                .weight_kg
                .to_string()
                .parse::<f64>()
                .unwrap_or_default(),
            priority,
            status,
            model.courier_id,
            model.zone,
            model.created_at,
            model.updated_at,
            model.assigned_at,
            model.delivered_at,
            model.not_delivered_reason,
            model.version as u32,
        ))
    }
}

fn priority_to_string(priority: Priority) -> String {
    match priority {
        Priority::Normal => "normal".to_string(),
        Priority::Urgent => "urgent".to_string(),
    }
}

fn string_to_priority(s: &str) -> Priority {
    match s.to_lowercase().as_str() {
        "urgent" => Priority::Urgent,
        _ => Priority::Normal,
    }
}

fn status_to_string(status: PackageStatus) -> String {
    match status {
        PackageStatus::Accepted => "accepted".to_string(),
        PackageStatus::InPool => "in_pool".to_string(),
        PackageStatus::Assigned => "assigned".to_string(),
        PackageStatus::InTransit => "in_transit".to_string(),
        PackageStatus::Delivered => "delivered".to_string(),
        PackageStatus::NotDelivered => "not_delivered".to_string(),
        PackageStatus::RequiresHandling => "requires_handling".to_string(),
    }
}

fn string_to_status(s: &str) -> PackageStatus {
    match s.to_lowercase().as_str() {
        "accepted" => PackageStatus::Accepted,
        "in_pool" => PackageStatus::InPool,
        "assigned" => PackageStatus::Assigned,
        "in_transit" => PackageStatus::InTransit,
        "delivered" => PackageStatus::Delivered,
        "not_delivered" => PackageStatus::NotDelivered,
        "requires_handling" => PackageStatus::RequiresHandling,
        _ => PackageStatus::Accepted,
    }
}
