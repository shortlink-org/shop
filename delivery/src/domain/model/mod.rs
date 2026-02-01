//! Domain Models
//!
//! This module contains all domain models including:
//! - Aggregates (Package, Courier)
//! - Entities (CourierLocation)
//! - Value Objects (Location, Address, etc.)
//! - Proto-generated models (delivery/)

pub mod courier;
pub mod courier_location;
pub mod package;
pub mod vo;

// Re-exports for convenience
pub use courier_location::{CourierLocation, CourierLocationError, LocationHistoryEntry, TimeRange};

// Proto-generated domain models organized as nested modules
pub mod domain {
    pub mod delivery {
        pub mod common {
            pub mod v1 {
                include!("domain.delivery.common.v1.rs");
            }
        }
        pub mod events {
            pub mod v1 {
                include!("domain.delivery.events.v1.rs");
            }
        }
        pub mod commands {
            pub mod v1 {
                include!("domain.delivery.commands.v1.rs");
            }
        }
        pub mod queries {
            pub mod v1 {
                include!("domain.delivery.queries.v1.rs");
            }
        }
    }
}
