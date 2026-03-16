//! Delivery Service Database Migrations
//!
//! Sea-ORM migrations for the delivery schema.

pub use sea_orm_migration::prelude::*;

mod m20260131_000001_create_couriers;
mod m20260131_000002_create_packages;
mod m20260131_000003_seed_couriers;
mod m20260131_000004_create_courier_location_history;
mod m20260205_000005_add_recipient_contacts_to_packages;
mod m20260311_000006_create_outbox_messages;
mod m20260311_000007_add_customer_phone_to_packages;
mod m20260316_000008_persist_courier_runtime_state;
mod m20260316_000009_drop_legacy_not_delivered_reason;

pub struct Migrator;

#[async_trait::async_trait]
impl MigratorTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![
            Box::new(m20260131_000001_create_couriers::Migration),
            Box::new(m20260131_000002_create_packages::Migration),
            Box::new(m20260131_000003_seed_couriers::Migration),
            Box::new(m20260131_000004_create_courier_location_history::Migration),
            Box::new(m20260205_000005_add_recipient_contacts_to_packages::Migration),
            Box::new(m20260311_000006_create_outbox_messages::Migration),
            Box::new(m20260311_000007_add_customer_phone_to_packages::Migration),
            Box::new(m20260316_000008_persist_courier_runtime_state::Migration),
            Box::new(m20260316_000009_drop_legacy_not_delivered_reason::Migration),
        ]
    }
}
