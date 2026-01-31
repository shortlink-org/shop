//! Courier Worker
//!
//! Handles courier-related workflows and activities.

pub mod activities;
pub mod workflow;

pub use activities::CourierActivities;
pub use workflow::*;

/// Task queue for courier workflows
pub const COURIER_TASK_QUEUE: &str = "COURIER_TASK_QUEUE";
