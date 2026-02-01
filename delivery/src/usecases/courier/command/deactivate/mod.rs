//! Deactivate Courier Command
//!
//! Deactivates a courier by setting their status to UNAVAILABLE.

mod command;
mod handler;

pub use command::Command;
pub use handler::{DeactivateCourierError, Handler, Response};
