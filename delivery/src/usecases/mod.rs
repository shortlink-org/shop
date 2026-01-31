//! Use Cases (Application Layer)
//!
//! Use cases orchestrate the flow of data and coordinate:
//! - Repository calls
//! - Domain services
//! - External service calls
//! - Transaction management
//!
//! Each use case follows the Command/Query + Handler pattern:
//! - Commands: modify state (command.rs + handler.rs)
//! - Queries: read-only operations (query.rs + handler.rs)
//!
//! Structure:
//! - courier/command/* - Courier write operations
//! - courier/query/* - Courier read operations
//! - package/command/* - Package/delivery write operations
//! - package/query/* - Package/delivery read operations

pub mod courier;
pub mod package;

// Re-export courier command types
pub use courier::command::{
    RegisterCommand, RegisterCourierError, RegisterHandler, RegisterResponse,
    UpdateLocationCommand, UpdateCourierLocationError, UpdateLocationHandler,
};

// Re-export courier query types
pub use courier::query::{
    CourierFilter, CourierWithState, GetCourierPoolError, GetPoolHandler as GetCourierPoolHandler,
    GetPoolQuery as GetCourierPoolQuery, GetPoolResponse as GetCourierPoolResponse,
};

// Re-export package command types
pub use package::command::{
    AcceptOrderCommand, AcceptOrderError, AcceptOrderHandler, AcceptOrderResponse,
    AssignOrderCommand, AssignOrderError, AssignOrderHandler, AssignOrderResponse,
    DeliverOrderCommand, DeliverOrderError, DeliverOrderHandler,
};

// Re-export package query types
pub use package::query::{
    GetPackagePoolError, GetPoolHandler as GetPackagePoolHandler,
    GetPoolQuery as GetPackagePoolQuery, GetPoolResponse as GetPackagePoolResponse, PackageFilter,
};
