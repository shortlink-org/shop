//! Sea-ORM Entity Definitions
//!
//! Database models for Sea-ORM that map to PostgreSQL tables.

pub mod courier;
pub mod location_history;
pub mod package;

pub use courier::Entity as CourierEntity;
pub use location_history::Entity as LocationHistoryEntity;
pub use package::Entity as PackageEntity;
