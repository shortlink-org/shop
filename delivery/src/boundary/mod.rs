//! Boundary Layer
//!
//! Contains port interfaces (traits) that define the contracts between
//! the application layer and infrastructure adapters.
//!
//! Following hexagonal architecture, ports are interfaces that:
//! - Are defined in the application/domain layer
//! - Are implemented by infrastructure adapters
//! - Allow swapping implementations without changing business logic

pub mod ports;
