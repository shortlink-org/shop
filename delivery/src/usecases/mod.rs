//! Use Cases (Application Layer)
//!
//! Use cases orchestrate the flow of data and coordinate:
//! - Repository calls
//! - Domain services
//! - External service calls
//! - Transaction management
//!
//! Each use case represents a single application operation.

pub mod accept_order;
pub mod assign_order;
pub mod deliver_order;
pub mod get_courier_pool;
pub mod get_package_pool;
pub mod register_courier;
pub mod update_courier_location;

// Re-export main use case types
pub use get_courier_pool::{CourierFilter, CourierWithState, GetCourierPoolError, GetCourierPoolResponse, GetCourierPoolUseCase};
pub use register_courier::{RegisterCourierError, RegisterCourierRequest, RegisterCourierResponse, RegisterCourierUseCase};
