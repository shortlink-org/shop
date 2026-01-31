//! Temporal Workers (Adapter Layer)
//!
//! Contains Temporal workflows and activities for the Delivery Service.
//! Workers are adapters that bridge Temporal orchestration with application use cases.
//!
//! Following hexagonal architecture:
//! - Workflows handle orchestration logic (deterministic)
//! - Activities are thin wrappers that call UseCases
//! - No business logic in this layer

pub mod courier;
pub mod delivery;

// Re-export task queues
pub use courier::COURIER_TASK_QUEUE;
pub use delivery::DELIVERY_TASK_QUEUE;
