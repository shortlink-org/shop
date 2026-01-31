//! Delivery Worker
//!
//! Handles delivery-related workflows (order assignment, delivery completion).

pub mod activities;
pub mod workflow;

pub use activities::DeliveryActivities;
pub use workflow::*;

/// Task queue for delivery workflows
pub const DELIVERY_TASK_QUEUE: &str = "DELIVERY_TASK_QUEUE";
